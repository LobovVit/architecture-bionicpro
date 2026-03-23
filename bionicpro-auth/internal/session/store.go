package session

import (
	"errors"
	"sync"
	"time"

	"bionicpro-auth/internal/models"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrPendingNotFound = errors.New("pending auth not found")
)

type Store struct {
	mu       sync.RWMutex
	sessions map[string]models.Session
	pending  map[string]models.PendingAuth
}

func NewStore() *Store {
	return &Store{
		sessions: make(map[string]models.Session),
		pending:  make(map[string]models.PendingAuth),
	}
}

func (s *Store) SavePending(p models.PendingAuth) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending[p.State] = p
}

func (s *Store) GetPending(state string) (models.PendingAuth, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.pending[state]
	if !ok {
		return models.PendingAuth{}, ErrPendingNotFound
	}
	return p, nil
}

func (s *Store) DeletePending(state string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.pending, state)
}

func (s *Store) SaveSession(sess models.Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sess.SessionID] = sess
}

func (s *Store) GetSession(sessionID string) (models.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[sessionID]
	if !ok {
		return models.Session{}, ErrSessionNotFound
	}
	return sess, nil
}

func (s *Store) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

func (s *Store) RotateSession(oldID, newID string) (models.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[oldID]
	if !ok {
		return models.Session{}, ErrSessionNotFound
	}
	delete(s.sessions, oldID)
	sess.SessionID = newID
	sess.UpdatedAt = time.Now().UTC()
	s.sessions[newID] = sess
	return sess, nil
}
