package controllers

import (
	"play"
	"play/samples/booking/app/models"
)

func addUser(c *play.Controller) play.Result {
	if user := connected(c); user != nil {
		c.RenderArgs["user"] = user
	}
	return nil
}

func connected(c *play.Controller) *models.User {
	if c.RenderArgs["user"] != nil {
		return c.RenderArgs["user"].(*models.User)
	}
	if username, ok := c.Session["user"]; ok {
		rows, err := c.Txn.Query(`
select UserId, Password, Name
  from User where Username = :Username`, username)
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		if !rows.Next() {
			return nil
		}

		user := &models.User{Username: username}
		err = rows.Scan(&user.UserId, &user.Password, &user.Name)
		if err != nil {
			panic(err)
		}
		return user
	}
	return nil
}

type Application struct {
	*play.Controller
}

func (c Application) Index() play.Result {
	if connected(c.Controller) != nil {
		return c.Redirect(Hotels.Index)
	}
	title := "Home"
	c.Flash.Error("Please log in first")
	return c.Render(title)
}

func (c Application) Register() play.Result {
	title := "Register"
	return c.Render(title)
}

func (c Application) SaveUser(user models.User, verifyPassword string) play.Result {
	c.Validation.Required(verifyPassword).Key("verifyPassword")
	c.Validation.Required(verifyPassword == user.Password).Key("verifyPassword").
		Message("Password does not match")
	user.Validate(c.Validation)

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Application.Register)
	}

	_, err := c.Txn.Exec("insert into User (Username, Password, Name) values (?, ?, ?)",
		user.Username, user.Password, user.Name)
	if err != nil {
		panic(err)
	}

	c.Session["user"] = user.Username
	c.Flash.Success("Welcome, " + user.Name)
	return c.Redirect(Hotels.Index)
}

func (c Application) Login(username, password string) play.Result {
	rows, err := c.Txn.Query(
		"select 1 from User where Username = ? and Password = ?",
		username, password)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	if rows.Next() {
		c.Session["user"] = username
		c.Flash.Success("Welcome, " + username)
		return c.Redirect(Hotels.Index)
	}

	c.Flash.Out["username"] = username
	c.Flash.Error("Login failed")
	return c.Redirect(Application.Index)
}

func (c Application) Logout() play.Result {
	for k := range c.Session {
		delete(c.Session, k)
	}
	return c.Redirect(Application.Index)
}

func init() {
	play.InterceptFunc(addUser, play.BEFORE, &Application{})
}
