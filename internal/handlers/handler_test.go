package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/suraj-kumal/very-short/internal/contracts"
)

const (
	testSiteURL     = "http://localhost:8080"
	testMixMultiple = 1
)

// ---------------------------------------------------------
// Fakes
// ---------------------------------------------------------

type fakeCache struct {
	data       map[string]string
	dirtyNodes []contracts.DirtyNode

	getCalled        bool
	putCalled        bool
	touchCalled      bool
	markCleanCalled  bool
	lastPutHash      string
	lastPutURL       string
	lastPutDirty     bool
	lastTouchedHash  string
	markedCleanNodes []contracts.DirtyNode
}

func newFakeCache() *fakeCache {
	return &fakeCache{
		data: make(map[string]string),
	}
}

func (f *fakeCache) Get(hash string) (string, bool) {
	f.getCalled = true

	url, ok := f.data[hash]
	return url, ok
}

func (f *fakeCache) Put(hash, url string, expire time.Time, dirty bool) {
	f.putCalled = true
	f.lastPutHash = hash
	f.lastPutURL = url
	f.lastPutDirty = dirty

	f.data[hash] = url
}

func (f *fakeCache) GetDirtyNodes() []contracts.DirtyNode {
	return f.dirtyNodes
}

func (f *fakeCache) MarkClean(nodes []contracts.DirtyNode) {
	f.markCleanCalled = true
	f.markedCleanNodes = nodes
}

func (f *fakeCache) Touch(hash string) {
	f.touchCalled = true
	f.lastTouchedHash = hash
}

// ---------------------------------------------------------
// Fake Concurrency Limiter
// ---------------------------------------------------------

type fakeConcurrencyLimiter struct {
	limitCalled bool
}

func (f *fakeConcurrencyLimiter) Limit(next http.Handler) http.Handler {
	f.limitCalled = true
	return next
}

// ---------------------------------------------------------
// Fake Last Sync
// ---------------------------------------------------------

type fakeLastSync struct {
	tryStartResult bool
	finishCalled   bool
}

func (f *fakeLastSync) Set(t time.Time) {}

func (f *fakeLastSync) Get() time.Time {
	return time.Time{}
}

func (f *fakeLastSync) CheckInterval(t time.Time) int {
	return 0
}

func (f *fakeLastSync) TryStartSync() bool {
	return f.tryStartResult
}

func (f *fakeLastSync) FinishSync() {
	f.finishCalled = true
}

// ---------------------------------------------------------
// Fake URL Store
// ---------------------------------------------------------

type fakeURLStore struct {
	beginTxFunc           func() (*sql.Tx, error)
	insertURLFunc         func(*sql.Tx, string) (int, error)
	updateHashFunc        func(*sql.Tx, int, string) error
	getURLFunc            func(string) (string, time.Time, error)
	updateAccessTimesFunc func([]contracts.DirtyNode) error
}

func (f *fakeURLStore) BeginTx() (*sql.Tx, error) {
	if f.beginTxFunc != nil {
		return f.beginTxFunc()
	}

	return nil, errors.New("not implemented")
}

func (f *fakeURLStore) InsertURLToDB(tx *sql.Tx, url string) (int, error) {
	if f.insertURLFunc != nil {
		return f.insertURLFunc(tx, url)
	}

	return 0, errors.New("not implemented")
}

func (f *fakeURLStore) UpdateHashToDB(
	tx *sql.Tx,
	id int,
	hash string,
) error {
	if f.updateHashFunc != nil {
		return f.updateHashFunc(tx, id, hash)
	}

	return errors.New("not implemented")
}

func (f *fakeURLStore) GetURLFromDB(
	hash string,
) (string, time.Time, error) {
	if f.getURLFunc != nil {
		return f.getURLFunc(hash)
	}

	return "", time.Time{}, errors.New("not implemented")
}

func (f *fakeURLStore) UpdateLastAccessTimes(
	nodes []contracts.DirtyNode,
) error {
	if f.updateAccessTimesFunc != nil {
		return f.updateAccessTimesFunc(nodes)
	}

	return errors.New("not implemented")
}

// ---------------------------------------------------------
// Helpers
// ---------------------------------------------------------

func newHandler(
	store URLStore,
	cache CacheStore,
	lastSync LastSync,
) *Handler {
	return HandlerStore(
		store,
		testSiteURL,
		testMixMultiple,
		cache,
		lastSync,
		&fakeConcurrencyLimiter{},
	)
}

func newSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db, mock
}

// ---------------------------------------------------------
// isValidURL
// ---------------------------------------------------------

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "valid http",
			url:  "http://example.com",
			want: true,
		},
		{
			name: "valid https",
			url:  "https://example.com",
			want: true,
		},
		{
			name: "invalid scheme",
			url:  "ftp://example.com",
			want: false,
		},
		{
			name: "invalid URL",
			url:  "not-a-url",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidURL(tt.url)

			if got != tt.want {
				t.Errorf(
					"isValidURL(%q) = %v, want %v",
					tt.url,
					got,
					tt.want,
				)
			}
		})
	}
}

// ---------------------------------------------------------
// CreateShortURL
// ---------------------------------------------------------

func TestCreateShortURL(t *testing.T) {
	tests := []struct {
		name       string
		form       string
		setupMock  func(sqlmock.Sqlmock)
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			form: "url=https%3A%2F%2Fexample.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectExec(
					`INSERT INTO url_data \(url\) VALUES \(\?\)`,
				).
					WithArgs("https://example.com").
					WillReturnResult(sqlmock.NewResult(10, 1))

				mock.ExpectExec(
					`UPDATE url_data SET hash = \? WHERE id = \?`,
				).
					WithArgs("A", 10).
					WillReturnResult(sqlmock.NewResult(0, 1))

				mock.ExpectCommit()
			},
			wantStatus: http.StatusCreated,
			wantBody:   "CREATED",
		},
		{
			name: "invalid url",
			form: "url=not-a-url",
			setupMock: func(mock sqlmock.Sqlmock) {
				// No database calls should happen.
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "URL must include http:// or https://",
		},
		{
			name: "begin transaction fails",
			form: "url=https%3A%2F%2Fexample.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().
					WillReturnError(errors.New("database unavailable"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Something went wrong.",
		},
		{
			name: "insert fails",
			form: "url=https%3A%2F%2Fexample.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectExec(
					`INSERT INTO url_data \(url\) VALUES \(\?\)`,
				).
					WithArgs("https://example.com").
					WillReturnError(errors.New("insert failed"))

				mock.ExpectRollback()
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Something went wrong.",
		},
		{
			name: "update hash fails",
			form: "url=https%3A%2F%2Fexample.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectExec(
					`INSERT INTO url_data \(url\) VALUES \(\?\)`,
				).
					WithArgs("https://example.com").
					WillReturnResult(sqlmock.NewResult(10, 1))

				mock.ExpectExec(
					`UPDATE url_data SET hash = \? WHERE id = \?`,
				).
					WithArgs("A", 10).
					WillReturnError(errors.New("update failed"))

				mock.ExpectRollback()
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Something went wrong.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newSQLMock(t)
			tt.setupMock(mock)

			store := &fakeURLStore{
				beginTxFunc: func() (*sql.Tx, error) {
					return db.Begin()
				},

				insertURLFunc: func(
					tx *sql.Tx,
					url string,
				) (int, error) {
					result, err := tx.Exec(
						"INSERT INTO url_data (url) VALUES (?)",
						url,
					)
					if err != nil {
						return 0, err
					}

					id, err := result.LastInsertId()
					return int(id), err
				},

				updateHashFunc: func(
					tx *sql.Tx,
					id int,
					hash string,
				) error {
					_, err := tx.Exec(
						"UPDATE url_data SET hash = ? WHERE id = ?",
						hash,
						id,
					)

					return err
				},
			}

			cache := newFakeCache()
			lastSync := &fakeLastSync{}

			h := newHandler(
				store,
				cache,
				lastSync,
			)

			req := httptest.NewRequest(
				http.MethodPost,
				"/shorten",
				strings.NewReader(tt.form),
			)

			req.Header.Set(
				"Content-Type",
				"application/x-www-form-urlencoded",
			)

			rec := httptest.NewRecorder()

			h.CreateShortURL(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf(
					"status = %d, want %d",
					rec.Code,
					tt.wantStatus,
				)
			}

			if !strings.Contains(
				rec.Body.String(),
				tt.wantBody,
			) {
				t.Errorf(
					"body does not contain %q: %s",
					tt.wantBody,
					rec.Body.String(),
				)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf(
					"unmet SQL expectations: %v",
					err,
				)
			}
		})
	}
}

// ---------------------------------------------------------
// RedirectURL - cache hit
// ---------------------------------------------------------

func TestRedirectURLCacheHit(t *testing.T) {
	cache := newFakeCache()
	cache.data["abc"] = "https://example.com"

	store := &fakeURLStore{}

	lastSync := &fakeLastSync{
		tryStartResult: false,
	}

	h := newHandler(
		store,
		cache,
		lastSync,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/abc",
		nil,
	)

	req.SetPathValue("hash", "abc")

	rec := httptest.NewRecorder()

	h.RedirectURL(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf(
			"status = %d, want %d",
			rec.Code,
			http.StatusFound,
		)
	}

	if location := rec.Header().Get("Location"); location != "https://example.com" {
		t.Errorf(
			"Location = %q, want %q",
			location,
			"https://example.com",
		)
	}

	if !cache.getCalled {
		t.Error("expected cache Get to be called")
	}
}

// ---------------------------------------------------------
// RedirectURL - cache miss / database hit
// ---------------------------------------------------------

func TestRedirectURLCacheMiss(t *testing.T) {
	cache := newFakeCache()

	expire := time.Now().Add(time.Hour)

	store := &fakeURLStore{
		getURLFunc: func(
			hash string,
		) (string, time.Time, error) {
			if hash != "abc" {
				t.Errorf(
					"hash = %q, want abc",
					hash,
				)
			}

			return "https://example.com", expire, nil
		},
	}

	lastSync := &fakeLastSync{
		tryStartResult: false,
	}

	h := newHandler(
		store,
		cache,
		lastSync,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/abc",
		nil,
	)

	req.SetPathValue("hash", "abc")

	rec := httptest.NewRecorder()

	h.RedirectURL(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf(
			"status = %d, want %d",
			rec.Code,
			http.StatusFound,
		)
	}

	if location := rec.Header().Get("Location"); location != "https://example.com" {
		t.Errorf(
			"Location = %q, want %q",
			location,
			"https://example.com",
		)
	}

	if !cache.putCalled {
		t.Error("expected cache Put to be called")
	}

	if cache.lastPutHash != "abc" {
		t.Errorf(
			"Put hash = %q, want abc",
			cache.lastPutHash,
		)
	}

	if !cache.touchCalled {
		t.Error("expected cache Touch to be called")
	}

	if cache.lastTouchedHash != "abc" {
		t.Errorf(
			"Touch hash = %q, want abc",
			cache.lastTouchedHash,
		)
	}
}

// ---------------------------------------------------------
// RedirectURL - not found
// ---------------------------------------------------------

func TestRedirectURLNotFound(t *testing.T) {
	cache := newFakeCache()

	store := &fakeURLStore{
		getURLFunc: func(
			hash string,
		) (string, time.Time, error) {
			return "", time.Time{}, sql.ErrNoRows
		},
	}

	lastSync := &fakeLastSync{
		tryStartResult: false,
	}

	h := newHandler(
		store,
		cache,
		lastSync,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/missing",
		nil,
	)

	req.SetPathValue("hash", "missing")

	rec := httptest.NewRecorder()

	h.RedirectURL(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf(
			"status = %d, want %d",
			rec.Code,
			http.StatusNotFound,
		)
	}
}

// ---------------------------------------------------------
// RedirectURL - database error
// ---------------------------------------------------------

func TestRedirectURLDatabaseError(t *testing.T) {
	cache := newFakeCache()

	store := &fakeURLStore{
		getURLFunc: func(
			hash string,
		) (string, time.Time, error) {
			return "", time.Time{}, errors.New("database failure")
		},
	}

	lastSync := &fakeLastSync{
		tryStartResult: false,
	}

	h := newHandler(
		store,
		cache,
		lastSync,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/abc",
		nil,
	)

	req.SetPathValue("hash", "abc")

	rec := httptest.NewRecorder()

	h.RedirectURL(rec, req)

	if rec.Code != http.StatusNotFound &&
		rec.Code != http.StatusInternalServerError {
		t.Errorf(
			"status = %d, want 404 or 500 depending on static/500.html",
			rec.Code,
		)
	}
}

// ---------------------------------------------------------
// RedirectURL - expired
// ---------------------------------------------------------

func TestRedirectURLExpired(t *testing.T) {
	cache := newFakeCache()

	store := &fakeURLStore{
		getURLFunc: func(
			hash string,
		) (string, time.Time, error) {
			return "https://example.com",
				time.Now().Add(-time.Hour),
				nil
		},
	}

	lastSync := &fakeLastSync{
		tryStartResult: false,
	}

	h := newHandler(
		store,
		cache,
		lastSync,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/abc",
		nil,
	)

	req.SetPathValue("hash", "abc")

	rec := httptest.NewRecorder()

	h.RedirectURL(rec, req)

	if rec.Code == http.StatusFound {
		t.Fatal("expired URL should not redirect")
	}
}

// ---------------------------------------------------------
// SyncLastAccessTime - no dirty nodes
// ---------------------------------------------------------

func TestSyncLastAccessTimeNoDirtyNodes(t *testing.T) {
	cache := newFakeCache()

	storeCalled := false

	store := &fakeURLStore{
		updateAccessTimesFunc: func(
			nodes []contracts.DirtyNode,
		) error {
			storeCalled = true
			return nil
		},
	}

	lastSync := &fakeLastSync{}

	h := newHandler(
		store,
		cache,
		lastSync,
	)

	h.SyncLastAccessTime()

	if storeCalled {
		t.Error(
			"expected database update not to be called",
		)
	}

	if cache.markCleanCalled {
		t.Error(
			"expected MarkClean not to be called",
		)
	}
}

// ---------------------------------------------------------
// SyncLastAccessTime - success
// ---------------------------------------------------------

func TestSyncLastAccessTimeSuccess(t *testing.T) {
	cache := newFakeCache()

	now := time.Now()

	cache.dirtyNodes = []contracts.DirtyNode{
		{
			Hash:           "abc",
			LastAccessTime: now,
		},
	}

	var gotNodes []contracts.DirtyNode

	store := &fakeURLStore{
		updateAccessTimesFunc: func(
			nodes []contracts.DirtyNode,
		) error {
			gotNodes = nodes
			return nil
		},
	}

	h := newHandler(
		store,
		cache,
		&fakeLastSync{},
	)

	h.SyncLastAccessTime()

	if len(gotNodes) != 1 {
		t.Fatalf(
			"got %d nodes, want 1",
			len(gotNodes),
		)
	}

	if gotNodes[0].Hash != "abc" {
		t.Errorf(
			"hash = %q, want %q",
			gotNodes[0].Hash,
			"abc",
		)
	}

	if !cache.markCleanCalled {
		t.Error(
			"expected MarkClean to be called",
		)
	}
}

// ---------------------------------------------------------
// SyncLastAccessTime - database failure
// ---------------------------------------------------------

func TestSyncLastAccessTimeDatabaseError(t *testing.T) {
	cache := newFakeCache()

	cache.dirtyNodes = []contracts.DirtyNode{
		{
			Hash:           "abc",
			LastAccessTime: time.Now(),
		},
	}

	store := &fakeURLStore{
		updateAccessTimesFunc: func(
			nodes []contracts.DirtyNode,
		) error {
			return errors.New("database failure")
		},
	}

	h := newHandler(
		store,
		cache,
		&fakeLastSync{},
	)

	h.SyncLastAccessTime()

	if cache.markCleanCalled {
		t.Error(
			"expected MarkClean not to be called after database failure",
		)
	}
}

// ---------------------------------------------------------
// CreateLongURL
// ---------------------------------------------------------

func TestCreateLongURL(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectBegin()

	mock.ExpectExec(
		`INSERT INTO url_data \(url\) VALUES \(\?\)`,
	).
		WithArgs("https://example.com").
		WillReturnResult(sqlmock.NewResult(10, 1))

	mock.ExpectExec(
		`UPDATE url_data SET hash = \? WHERE id = \?`,
	).
		WithArgs("A", 10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	store := &fakeURLStore{
		beginTxFunc: func() (*sql.Tx, error) {
			return db.Begin()
		},

		insertURLFunc: func(
			tx *sql.Tx,
			url string,
		) (int, error) {
			result, err := tx.Exec(
				"INSERT INTO url_data (url) VALUES (?)",
				url,
			)
			if err != nil {
				return 0, err
			}

			id, err := result.LastInsertId()

			return int(id), err
		},

		updateHashFunc: func(
			tx *sql.Tx,
			id int,
			hash string,
		) error {
			_, err := tx.Exec(
				"UPDATE url_data SET hash = ? WHERE id = ?",
				hash,
				id,
			)

			return err
		},
	}

	h := newHandler(
		store,
		newFakeCache(),
		&fakeLastSync{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/long",
		strings.NewReader(
			"url=https%3A%2F%2Fexample.com",
		),
	)

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	rec := httptest.NewRecorder()

	h.CreateLongURL(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf(
			"status = %d, want %d",
			rec.Code,
			http.StatusCreated,
		)
	}

	if !strings.Contains(
		rec.Body.String(),
		"CREATED",
	) {
		t.Error("expected CREATED in response")
	}

	if !strings.Contains(
		rec.Body.String(),
		"/A",
	) {
		t.Error("expected generated hash A in response")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(
			"unmet SQL expectations: %v",
			err,
		)
	}
}

// ---------------------------------------------------------
// ConcurrencyLimiter
// ---------------------------------------------------------

func TestFakeConcurrencyLimiter(t *testing.T) {
	limiter := &fakeConcurrencyLimiter{}

	called := false

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := limiter.Limit(next)

	if !limiter.limitCalled {
		t.Error("expected Limit to be called")
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected wrapped handler to be called")
	}

	if rec.Code != http.StatusOK {
		t.Errorf(
			"status = %d, want %d",
			rec.Code,
			http.StatusOK,
		)
	}
}
