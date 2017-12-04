// Copyright 2017 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mail

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
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
	log.Printf("%s %s", req.Method, req.URL)
	var err error
	switch {
	case req.URL.Path == "/",
		req.URL.Path == "/index.html":
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(mailHtml))
	case req.URL.Path == "/mailboxes/":
		// URL: /mailboxes/
		var folders []string
		for f := range s.mr.BoxPaths {
			folders = append(folders, f)
		}
		sort.Strings(folders)
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(folders)
	case strings.HasPrefix(req.URL.Path, "/mailbox/"):
		// URL: /mailbox/$MAILBOX
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/mailbox/")
		err = s.Folder(w, req)
	case strings.HasPrefix(req.URL.Path, "/message/"):
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/message/")
		// URL: /message/$MAILBOX/$MSGID
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
	folder := req.URL.Path
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
	// Path = "$FOLDER/$MSGID"
	parts := strings.Split(req.URL.Path, "/")
	if len(parts) > 2 {
		http.NotFound(w, req)
		return nil
	}
	folder := parts[0]
	idx, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "idx "+parts[1]+" is not a number",
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
