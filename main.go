package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Server struct {
	port              int
	t                 time.Time
	jobStart          time.Time
	waitStartupTime   time.Duration
	waitLivenessTime  time.Duration
	waitReadinessTime time.Duration
	jobDuration       time.Duration
	isReadinessEqualLiveness bool
	maxReadinessCount int
	currReadinessCount int
}

func main() {
	var s Server
	s.port = 8080

	err := s.getEnvValues()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/startupProbe", s.startupProbe)
	http.HandleFunc("/livenessProbe", s.livenessProbe)
	http.HandleFunc("/readinessProbe", s.readinessProbe)
	http.HandleFunc("/maxReadinessCountProbe", s.maxReadinessCountProbe)
	http.HandleFunc("/startJob", s.startJob)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Start time
	s.t = time.Now()
	s.currReadinessCount = 0

	fmt.Printf("Starting server. Listening on port: %d\n", s.port)
	log.Fatal(srv.ListenAndServe())
}

func getEnvToDuration(e string) (d time.Duration, err error) {
	var envValue int
	envValue, err = strconv.Atoi(os.Getenv(e))
	d = time.Duration(envValue) * time.Second
	return
}

func getEnvToInt(e string) (int, error) {
	envValue, err := strconv.Atoi(os.Getenv(e))
	if err == nil {
		return envValue, nil
	}
	return -1, err
}

func getEnvToBool(e string) (bool, error) {
	boolVal, err := strconv.ParseBool(os.Getenv(e))
	if err == nil {
		return boolVal, nil
	}
	return false, err
}

func (s *Server) getEnvValues() (err error) {
	s.waitStartupTime, err = getEnvToDuration("WAIT_STARTUP_TIME")
	if err != nil {
		return
	}
	s.waitLivenessTime, err = getEnvToDuration("WAIT_LIVENESS_TIME")
	if err != nil {
		return
	}
	s.waitReadinessTime, err = getEnvToDuration("WAIT_READINESS_TIME")
	if err != nil {
		return
	}
	s.jobDuration, err = getEnvToDuration("JOB_DURATION_TIME")
	if err != nil {
		return
	}

	s.isReadinessEqualLiveness, err = getEnvToBool("IS_READINESS_EQUALS_LIVENESS")
	log.Println("isReadinessEqualLiveness: ", s.isReadinessEqualLiveness)

	s.maxReadinessCount, err = getEnvToInt("MAX_READINESS_COUNT")
	if err != nil {
		return
	}
	log.Println("maxReadinessCount: ", s.maxReadinessCount)

	return
}

func (s *Server) startupProbe(w http.ResponseWriter, r *http.Request) {
	if time.Since(s.t) > s.waitStartupTime {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(503)
	}
}

func (s *Server) livenessProbe(w http.ResponseWriter, r *http.Request) {
	if time.Since(s.t) > s.waitLivenessTime {
		w.WriteHeader(200)
		log.Println("liveness 200 OK")
	} else {
		w.WriteHeader(503)
		log.Println("liveness 503 NOK")
	}
}

func (s *Server) maxReadinessCountProbe(w http.ResponseWriter, r *http.Request) {
	if s.currReadinessCount >= s.maxReadinessCount {
		w.WriteHeader(503)
		log.Println("currReadinessCount >= s.maxReadinessCount 503 NOK")
	} else {
		s.currReadinessCount++
		w.WriteHeader(200)
		log.Println("currReadinessCount < s.maxReadinessCount 200 OK")
	}
}

func (s *Server) readinessProbe(w http.ResponseWriter, r *http.Request) {
	if s.isReadinessEqualLiveness && (time.Since(s.t) > s.waitLivenessTime) {
		w.WriteHeader(200)
		log.Println("readiness == liveness 200 OK")
	} else if time.Since(s.t) > s.waitReadinessTime && time.Since(s.jobStart) > s.jobDuration {
		w.WriteHeader(200)
		log.Println("readiness > jobduration 200 OK")
	} else {
		w.WriteHeader(503)
		log.Println("readiness 503 NOK")
	}
}

func (s *Server) startJob(w http.ResponseWriter, r *http.Request) {
	if time.Since(s.jobStart) > s.jobDuration {
		s.jobStart = time.Now()
		fmt.Fprintf(w, "Pod (%s)\nStarting job. Unavailable till: %s", os.Getenv("HOSTNAME"), s.jobStart.Add(s.jobDuration).Format("Mon Jan _2 15:04:05 2006"))
	} else {
		fmt.Fprintf(w, "Still running job. Unavailable till: %v", s.jobStart.Add(s.jobDuration))
	}
}
