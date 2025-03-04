package protocol

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/bigwhite/go-cache-prog/cache"
)

// Cmd, Request, and Response structs
type Cmd string

const (
	CmdPut   Cmd = "put"
	CmdGet   Cmd = "get"
	CmdClose Cmd = "close"
)

type Request struct {
	ID       int64
	Command  Cmd
	ActionID []byte    `json:",omitempty"`
	OutputID []byte    `json:",omitempty"`
	Body     io.Reader `json:"-"`
	BodySize int64     `json:",omitempty"`
	ObjectID []byte    `json:",omitempty"` // Will be Deprecated in go 1.25 and later version
}

type Response struct {
	ID            int64      `json:",omitempty"`
	Err           string     `json:",omitempty"`
	KnownCommands []Cmd      `json:",omitempty"`
	Miss          bool       `json:",omitempty"`
	OutputID      []byte     `json:",omitempty"`
	Size          int64      `json:",omitempty"`
	Time          *time.Time `json:",omitempty"`
	DiskPath      string     `json:",omitempty"`
}

type RequestHandler struct {
	reader        *bufio.Reader
	writer        io.Writer
	cache         *cache.Cache
	verbose       bool
	gets          int //statistics
	getMiss       int
	nextRequestID int64
}

func NewRequestHandler(reader *bufio.Reader, writer io.Writer, cache *cache.Cache, verbose bool) *RequestHandler {
	return &RequestHandler{
		reader:        reader,
		writer:        writer,
		cache:         cache,
		verbose:       verbose,
		gets:          0,
		getMiss:       0,
		nextRequestID: 1, //Start from 1 to match original logic where ID 0 is for init
	}
}

func (rh *RequestHandler) HandleRequests() error {
	for {
		req, err := rh.readRequest()
		if err != nil {
			if err == io.EOF {
				return nil // Clean exit on EOF
			}
			return fmt.Errorf("error reading request: %w", err)
		}
		if req == nil { // empty line
			continue
		}

		rh.processRequest(req)
	}
}

func (rh *RequestHandler) readRequest() (*Request, error) {
	line, err := rh.reader.ReadBytes('\n')
	if err != nil {
		return nil, err // Could be io.EOF, which is handled above
	}

	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return nil, nil // Skip empty lines
	}

	var req Request
	if err := json.Unmarshal(line, &req); err != nil {
		// Check for base64 encoded body (part of a PUT request)
		if len(line) >= 2 && line[0] == '"' && line[len(line)-1] == '"' {
			if rh.verbose {
				log.Println("Found base64 body line")
			}
			return nil, fmt.Errorf("base64 body outside of put context not supported")
		}

		return nil, fmt.Errorf("failed to unmarshal request: %w, input: %q", err, line)
	}

	return &req, nil
}

func (rh *RequestHandler) processRequest(req *Request) {
	if rh.verbose {
		log.Printf("Received request: ID=%d, Command=%s, ActionID=%x, OutputID=%x, BodySize=%d\n", req.ID,
			req.Command, req.ActionID, req.OutputID, req.BodySize)
	}

	switch req.Command {
	case CmdPut:
		rh.handlePut(req)
	case CmdGet:
		rh.handleGet(req)
	case CmdClose:
		rh.handleClose(req)
	default:
		rh.sendErrorResponse(req.ID, fmt.Sprintf("unknown command: %s", req.Command))
	}
}

func (rh *RequestHandler) handlePut(req *Request) {

	var bodyData []byte
	// For bodysize > 0, read the next line(s) for the Base64 body.
	if req.BodySize > 0 {

		bodyLine, err := rh.reader.ReadBytes('\n')
		if len(bodyLine) != 1 {
			log.Printf("Put request: ID=%d, Extra newline missed\n", req.ID)
		}

		bodyLine, err = rh.reader.ReadBytes('\n')
		if err != nil {
			rh.sendErrorResponse(req.ID, fmt.Sprintf("failed to read body: %v", err))
			return
		}

		if rh.verbose {
			log.Printf("Put request: ID=%d, Actual BodyLen=%d\n", req.ID, len(bodyLine))
		}

		bodyLine = bytes.TrimSpace(bodyLine)
		if len(bodyLine) < 2 || bodyLine[0] != '"' || bodyLine[len(bodyLine)-1] != '"' {
			rh.sendErrorResponse(req.ID, "malformed body format")
			return
		}

		base64Body := bodyLine[1 : len(bodyLine)-1]
		bodyData, err = base64.StdEncoding.DecodeString(string(base64Body))
		if err != nil {
			rh.sendErrorResponse(req.ID, fmt.Sprintf("failed to decode body: %v", err))
			return
		}

		if int64(len(bodyData)) != req.BodySize {
			rh.sendErrorResponse(req.ID, "body size mismatch")
			return
		}
	}

	diskPath, err := rh.cache.Put(req.ActionID, req.OutputID, bodyData, req.BodySize)
	if err != nil {
		rh.sendErrorResponse(req.ID, fmt.Sprintf("put failed: %v", err))
		return
	}

	resp := Response{
		ID:       req.ID,
		DiskPath: diskPath,
	}
	rh.sendResponse(resp)
}

func (rh *RequestHandler) handleGet(req *Request) {
	rh.gets++
	entry, found, err := rh.cache.Get(req.ActionID)
	if err != nil {
		rh.sendErrorResponse(req.ID, fmt.Sprintf("get failed: %v", err))
		return
	}

	if !found {
		rh.getMiss++
		resp := Response{
			ID:   req.ID,
			Miss: true,
		}
		rh.sendResponse(resp)
		return
	}

	resp := Response{
		ID:       req.ID,
		OutputID: entry.OutputID,
		Size:     entry.Size,
		Time:     &entry.Time,
		DiskPath: entry.DiskPath,
	}
	rh.sendResponse(resp)
}

func (rh *RequestHandler) handleClose(req *Request) {
	resp := Response{ID: req.ID}
	rh.sendResponse(resp)
	// No explicit return here, the loop in HandleRequests will exit due to EOF on next read attempt
}

func (rh *RequestHandler) sendResponse(resp Response) {
	if err := json.NewEncoder(rh.writer).Encode(resp); err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}

func (rh *RequestHandler) sendErrorResponse(reqID int64, errMsg string) {
	resp := Response{
		ID:  reqID,
		Err: errMsg,
	}
	rh.sendResponse(resp)
}

func (rh *RequestHandler) Stats() (int, int) {
	return rh.gets, rh.getMiss
}
