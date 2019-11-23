package model

import (
	"fmt"
	"time"

	"github.com/google/go-github/v28/github"
	"github.com/naiba/com"
)

// User ...
type User struct {
	Common    `json:"common,omitempty"`
	Login     string `json:"login,omitempty"`      // 登录名
	AvatarURL string `json:"avatar_url,omitempty"` // 头像地址
	Name      string `json:"name,omitempty"`       // 昵称
	Blog      string `json:"blog,omitempty"`       // 网站链接
	Email     string `json:"email,omitempty"`      // 邮箱
	Hireable  bool   `json:"hireable,omitempty"`
	Bio       string `json:"bio,omitempty"` // 个人简介

	Token        string    `gorm:"UNIQUE_INDEX" json:"-"`   // 认证 Token
	TokenExpired time.Time `json:"token_expired,omitempty"` // Token 过期时间
	SuperAdmin   bool      `json:"super_admin,omitempty"`   // 超级管理员
}

// NewUserFromGitHub ..
func NewUserFromGitHub(gu *github.User) User {
	var u User
	u.ID = uint64(gu.GetID())
	u.Login = gu.GetLogin()
	u.AvatarURL = gu.GetAvatarURL()
	u.Name = gu.GetName()
	u.Blog = gu.GetBlog()
	u.Blog = gu.GetBlog()
	u.Email = gu.GetEmail()
	u.Hireable = gu.GetHireable()
	u.Bio = gu.GetBio()
	return u
}

// IssueNewToken ...
func (u *User) IssueNewToken() {
	u.Token = com.MD5(fmt.Sprintf("%s%d%s", time.Now(), u.ID, u.Login))
	u.TokenExpired = time.Now().AddDate(0, 0, 14)
}
