package handler

import (
	"asset-management-api/internal/middleware"
	"asset-management-api/internal/service/interfaces"
	"asset-management-api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ManagerHandler struct {
	managerService interfaces.ManagerService
}

func NewManagerHandler(managerService interfaces.ManagerService) *ManagerHandler {
	return &ManagerHandler{managerService: managerService}
}

// GET /teams/:teamId/assets
func (h *ManagerHandler) GetTeamAssets(c *gin.Context) {
	managerID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID format", err)
		return
	}

	assets, err := h.managerService.GetTeamAssets(teamID, managerID)
	if err != nil {
		if err.Error() == "access denied: only managers can view team assets" {
			utils.ForbiddenResponse(c, "Manager role required")
			return
		}
		if err.Error() == "access denied: you are not a manager of this team" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		if err.Error() == "team not found: record not found" {
			utils.NotFoundResponse(c, "Team not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get team assets", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Team assets retrieved successfully", assets)
}

// GET /users/:userId/assets
func (h *ManagerHandler) GetUserAssets(c *gin.Context) {
	managerID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	targetUserIDStr := c.Param("userId")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	assets, err := h.managerService.GetUserAssets(targetUserID, managerID)
	if err != nil {
		if err.Error() == "access denied: only managers can view user assets" {
			utils.ForbiddenResponse(c, "Manager role required")
			return
		}
		if err.Error() == "access denied: you can only view assets of users in your teams" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get user assets", err)
		return
	}

	utils.SuccessResponse(c,http.StatusOK, "Team assets retrieved successfully", assets)
}