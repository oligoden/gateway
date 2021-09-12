package session

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/oligoden/chassis"
	"github.com/oligoden/chassis/device/model"
	"github.com/oligoden/chassis/device/model/data"
	"github.com/oligoden/chassis/storage/gosql"
)

type Model struct {
	model.Default
	user    string
	session string
}

func NewModel(r *http.Request, s model.Connector) *Model {
	m := &Model{}
	m.Default = model.Default{}
	m.Request = r
	m.Store = s
	m.NewData = func() data.Operator { return NewRecord() }
	m.Data(NewRecord())
	return m
}

func (m *Model) Bind() {
	if m.Err() != nil {
		return
	}

	m.user = m.Request.Header.Get("X_user")
	sessionCookie, err := m.Request.Cookie("session")
	if err != nil {
		if err != http.ErrNoCookie {
			m.Err("internal error, could not get cookie, %w", err)
			return
		}
	}

	if sessionCookie == nil {
		m.session = ""
	} else {
		m.session = sessionCookie.Value
	}
}

func (m *Model) Authenticate() {
	if m.Err() != nil {
		return
	}

	if m.session == "" {
		fmt.Println("authentication: no session cookie value, creating new session cookie")
		m.createSession()
		return
	}

	c := m.Store.Connect(m.User())

	where := gosql.NewWhere("sessions.hash = ?", m.session)
	c.AddModifiers(where)
	session := NewRecord()
	c.Read(session)
	if c.Err() != nil {
		m.Err(chassis.Mark("reading session", c.Err()))
		return
	}

	if session.ID == 0 {
		fmt.Println("authentication: unknown session cookie value, creating new session cookie")
		m.createSession()
		return
	}

	fmt.Println("got session", session.ID)
	m.Request.Header.Set("X_session", fmt.Sprint(session.ID))

	userRecords := gosql.UserRecords{}
	joinSessionUsers := gosql.NewJoin("LEFT JOIN session_users ON session_users.user_id = users.owner_id")
	joinSessions := gosql.NewJoin("LEFT JOIN sessions ON sessions.id = session_users.session_id")
	c.AddModifiers(joinSessionUsers, joinSessions, where)
	c.Read(&userRecords)
	if c.Err() != nil {
		m.Err(chassis.Mark("reading users", c.Err()))
		return
	}

	users := []gosql.User(userRecords)
	if len(users) == 0 {
		m.Request.Header.Set("X_user", "0")
		return
	}

	if m.user == "" {
		m.user = users[0].Username
		fmt.Println("using user", users[0].OwnerID)
		m.Request.Header.Set("X_user", fmt.Sprint(users[0].OwnerID))
		return
	}

	for _, user := range users {
		if user.Username == m.user {
			fmt.Println("using user", user.OwnerID)
			m.Request.Header.Set("X_user", fmt.Sprint(user.OwnerID))
			return
		}
	}

	m.user = users[0].Username
	fmt.Println("using user", users[0].OwnerID)
	m.Request.Header.Set("X_user", fmt.Sprint(users[0].OwnerID))
}

func (m *Model) createSession() {
	e := NewRecord()
	e.IP = m.Request.RemoteAddr
	e.UserAgent = m.Request.UserAgent()
	m.Data(e)
	m.Create()
	m.session = e.Hash
	m.Request.Header.Set("X_user", "0")
	m.Request.Header.Set("X_session", fmt.Sprint(e.ID))
}

func (m *Model) CreateUser() {
	if m.Err() != nil {
		return
	}

	if m.Request.Method != http.MethodPost {
		return
	}

	m.BindUser()
	m.Bind()

	e := gosql.NewUserRecord(m.Request, m.Store.Rnd())
	m.Data(e)

	buf, _ := ioutil.ReadAll(m.Request.Body)
	m.Request.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	m.Default.Bind()
	m.Request.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

	err := e.Prepare()
	if err != nil {
		m.Err(chassis.Mark("preparing user for create", err))
		return
	}

	c := m.Store.Connect(m.User())
	c.Create(e)
	if c.Err() != nil {
		m.Err(chassis.Mark("creating user", c.Err()))
		return
	}
	fmt.Println("created user", e.IDValue())

	err = e.Hasher()
	if err != nil {
		m.Err(chassis.Mark("hashing user record", err))
		return
	}

	c = m.Store.Connect(e.IDValue(), []uint{})
	where := gosql.NewWhere("owner_id = ?", e.IDValue())
	c.AddModifiers(where)
	c.Update(e)
	if c.Err() != nil {
		m.Err(chassis.Mark("updating user with hash", c.Err()))
		return
	}

	err = e.Complete()
	if err != nil {
		m.Err(chassis.Mark("completing user record", err))
		return
	}

	eSessionUser := NewSessionUsersRecord()
	eSessionUser.SessionID = m.Session()
	eSessionUser.UserID = e.IDValue()
	m.Data(eSessionUser)

	c.Create(eSessionUser)
	if c.Err() != nil {
		m.Err(chassis.Mark("creating session user record", c.Err()))
		return
	}

	err = eSessionUser.Hasher()
	if err != nil {
		m.Err(chassis.Mark("hashing session user record", err))
		return
	}

	c.Update(eSessionUser)
	if c.Err() != nil {
		m.Err(chassis.Mark("updating session user with hash", c.Err()))
		return
	}
}

func (m *Model) Signin() {
	if m.Err() != nil {
		return
	}

	m.BindUser()
	user, _ := m.User()
	if user != 0 {
		m.Err("bad request, already signed in")
		return
	}

	m.Request.ParseForm()

	username := m.Request.FormValue("username")
	if username == "" {
		m.Err("bad request, empty username")
		return
	}

	password := m.Request.FormValue("password")
	if password == "" {
		m.Err("bad request, empty password")
		return
	}

	c := m.Store.Connect(m.User())

	// eProfile := profile.NewRecord()
	// m.Data(eProfile)
	join := gosql.NewJoin("JOIN users ON profiles.owner_id = users.owner_id")
	where := gosql.NewWhere("email=? OR username=?", username, username)
	// c.AddModifiers(join, where)
	// c.Read(eProfile)
	// if c.Err() != nil {
	// 	m.Err(c.Err())
	// 	return
	// }

	eUser := &gosql.User{}
	join = gosql.NewJoin("JOIN profiles ON profiles.owner_id = users.owner_id")
	c.AddModifiers(join, where)
	c.Read(eUser)
	if c.Err() != nil {
		m.Err(c.Err())
		return
	}

	if eUser.IDValue() == 0 {
		m.Err("bad request, incorrect credentials")
		return
	}

	h := sha256.New()
	h.Write([]byte(password + eUser.Salt))
	bs := h.Sum(nil)

	if eUser.PassHash != fmt.Sprintf("%x", bs) {
		m.Err("bad request, incorrect credentials")
		return
	}

	eSessionUser := NewSessionUsersRecord()
	eSessionUser.SessionID = m.Session()
	eSessionUser.UserID = eUser.OwnerID
	c.Create(eSessionUser)
	if c.Err() != nil {
		m.Err(c.Err())
		return
	}

	m.user = eUser.Username
	m.Request.Header.Set("X_user", fmt.Sprint(eUser.OwnerID))
}

func (m *Model) Signout() {
	if m.Err() != nil {
		return
	}

	m.BindUser()
	c := m.Store.Connect(m.User())

	m.Bind()
	where := gosql.NewWhere("hash=?", m.session)
	c.AddModifiers(where)
	eSesh := NewRecord()
	c.Read(eSesh)
	if c.Err() != nil {
		m.Err(c.Err)
		return
	}

	eUser := &gosql.User{}
	user, _ := m.User()
	where = gosql.NewWhere("owner_id=?", user)
	c.AddModifiers(where)
	c.Read(eUser)
	if c.Err() != nil {
		m.Err(c.Err())
		return
	}

	eSessionUser := NewSessionUsersRecord()
	where = gosql.NewWhere("session_id=? AND user_id=?", eSesh.ID, eUser.OwnerID)
	c.Delete(eSessionUser)
	if c.Err() != nil {
		m.Err(c.Err())
		return
	}

	m.user = ""
}
