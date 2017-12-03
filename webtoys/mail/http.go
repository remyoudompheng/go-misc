// Copyright 2017 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mail

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	mr MailReader
}

func NewServer(paths map[string]string) *Server {
	srv := &Server{
		mr: MailReader{
			BoxPaths: paths,
			Boxes:    make(map[string]*Mailbox),
		},
	}
	return srv
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var err error
	switch {
	case strings.HasPrefix(req.URL.Path, "/folder"):
		err = s.Folder(w, req)
	case strings.HasPrefix(req.URL.Path, "/message"):
		err = s.Message(w, req)
	default:
		http.NotFound(w, req)
	}
	if err != nil {
		log.Printf("could not send response: %s", err)
	}
}

func (s *Server) Folder(w http.ResponseWriter, req *http.Request) error {
	req.ParseForm()
	folder := req.FormValue("folder")
	idxS := req.FormValue("idx")
	idx := 0
	if idxS != "" {
		n, err := strconv.Atoi(idxS)
		if err != nil {
			http.Error(w, "idx "+idxS+" is not a number",
				http.StatusBadRequest)
			return nil
		}
		idx = n
	}
	hdrs, err := s.mr.ListFolder(folder, idx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(hdrs)
}

func (s *Server) Message(w http.ResponseWriter, req *http.Request) error {
	req.ParseForm()
	folder := req.FormValue("folder")
	idxS := req.FormValue("idx")
	idx, err := strconv.Atoi(idxS)
	if err != nil {
		http.Error(w, "idx "+idxS+" is not a number",
			http.StatusBadRequest)
		return nil
	}
	msg, err := s.mr.Message(folder, idx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(msg)
}
