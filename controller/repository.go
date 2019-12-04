package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	GitHubAPI "github.com/google/go-github/v28/github"
	"github.com/jinzhu/gorm"

	"github.com/naiba/poorsquad/model"
	"github.com/naiba/poorsquad/service/dao"
	"github.com/naiba/poorsquad/service/github"
)

// RepositoryController ..
type RepositoryController struct {
}

// ServeRepository ..
func ServeRepository(r gin.IRoutes) {
	rc := RepositoryController{}
	r.POST("/repository", rc.addOrEdit)
	r.DELETE("/repository/:id/:name", rc.delete)
}

type repositoryForm struct {
	ID        uint64 `json:"id,omitempty"`
	Name      string `binding:"required" json:"name,omitempty"`
	AccountID uint64 `binding:"required" json:"account_id,omitempty"`
	Private   string `binding:"required" json:"private,omitempty"`
}

func (rc *RepositoryController) addOrEdit(c *gin.Context) {
	var rf repositoryForm
	if err := c.ShouldBindJSON(&rf); err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("格式错误：%s", err),
		})
		return
	}
	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)

	// 验证管理权限
	var distAccount model.Account
	if err := dao.DB.First(&distAccount, "id = ?", rf.AccountID).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("请求受限：%s", err),
		})
		return
	}

	var comp model.Company
	comp.ID = distAccount.CompanyID
	if _, err := comp.CheckUserPermission(dao.DB, u.ID, model.UCPSuperManager); err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	var err error
	var repostory model.Repository
	var repo GitHubAPI.Repository
	if rf.ID != 0 {
		if rf.AccountID != distAccount.ID {
			err = errors.New("GitHub 尚未完善账户间转移 API")
		}
		if err == nil {
			repostory.ID = rf.ID
			err = dao.DB.First(&repostory).Error
		}
	}
	// 添加仓库
	ctx := context.Background()
	client := github.NewAPIClient(ctx, distAccount.Token)
	repo.Name = &rf.Name
	private := rf.Private == "on"
	repo.Private = &private
	var resp *GitHubAPI.Repository
	if rf.ID != 0 {
		resp, _, err = client.Repositories.Edit(ctx, distAccount.Login, repostory.Name, &repo)
	} else {
		resp, _, err = client.Repositories.Create(ctx, "", &repo)
	}
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("GitHub 同步：%s", err),
		})
		return
	}
	r := model.NewRepositoryFromGitHub(resp)
	r.AccountID = distAccount.ID
	r.SyncedAt = time.Now()
	if err := dao.DB.Save(&r).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("数据库错误：%s", err),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: http.StatusOK,
	})
}

func (rc *RepositoryController) delete(c *gin.Context) {
	u := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)

	// 验证管理权限
	var repo model.Repository
	var account model.Account
	var comp model.Company
	err := dao.DB.First(&repo, "id = ?", c.Param("id")).Error

	if err == nil {
		if repo.Name != c.Param("name") {
			c.JSON(http.StatusOK, model.Response{
				Code:    http.StatusBadRequest,
				Message: "仓库名称不匹配",
			})
			return
		}
		err = dao.DB.First(&account, "id = ?", repo.AccountID).Error
	}

	if err == nil {
		comp.ID = account.CompanyID
		_, err = comp.CheckUserPermission(dao.DB, u.ID, model.UCPSuperManager)
	}

	var tx *gorm.DB
	if err == nil {
		tx = dao.DB.Begin()
		err = tx.Delete(model.UserRepository{}, "repository_id = ?", repo.ID).Error
	}
	if err == nil {
		err = tx.Delete(repo).Error
	}
	if err == nil {
		ctx := context.Background()
		client := github.NewAPIClient(ctx, account.Token)
		_, err = client.Repositories.Delete(ctx, account.Login, repo.Name)
	}
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("出现错误：%s", err),
		})
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, model.Response{
		Code: http.StatusOK,
	})
}
