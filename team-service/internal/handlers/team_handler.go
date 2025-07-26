// internal/handlers/team_handler.go
package handlers

import (
	"net/http"
	"team-service/internal/models"
	"team-service/internal/services"
	"team-service/internal/utils"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TeamHandler struct {
	teamService services.TeamService
}

func NewTeamHandler(teamService services.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

// POST /teams
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	
	var req models.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	team, err := h.teamService.CreateTeam(userID.(uuid.UUID), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	
	utils.SuccessResponse(c, http.StatusCreated, "Team created successfully", team)
}

// GET /teams/:teamId
func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}
	
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	
	team, err := h.teamService.GetTeam(teamID, userID.(uuid.UUID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}
	
	utils.SuccessResponse(c, http.StatusOK, "Team retrieved successfully", team)
}

// GET /teams
func (h *TeamHandler) GetUserTeams(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	
	teams, err := h.teamService.GetUserTeams(userID.(uuid.UUID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	
	utils.SuccessResponse(c, http.StatusOK, "Teams retrieved successfully", teams)
}

// POST /teams/:teamId/members
func (h *TeamHandler) AddMember(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}
	
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	
	var req models.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	memberID, err := uuid.Parse(req.UserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}
	
	err = h.teamService.AddMember(teamID, userID.(uuid.UUID), memberID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	
	utils.SuccessResponse(c, http.StatusOK, "Member added successfully", nil)
}

// DELETE /teams/:teamId/members/:memberId
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}
	
	memberIDStr := c.Param("memberId")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid member ID")
		return
	}
	
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	
	err = h.teamService.RemoveMember(teamID, userID.(uuid.UUID), memberID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	
	utils.SuccessResponse(c, http.StatusOK, "Member removed successfully", nil)
}

// POST /teams/:teamId/managers
func (h *TeamHandler) AddManager(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}
	
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	
	var req models.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	managerID, err := uuid.Parse(req.UserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}
	
	err = h.teamService.AddManager(teamID, userID.(uuid.UUID), managerID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	
	utils.SuccessResponse(c, http.StatusOK, "Manager added successfully", nil)
}

// DELETE /teams/:teamId/managers/:managerId
func (h *TeamHandler) RemoveManager(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}
	
	managerIDStr := c.Param("managerId")
	managerID, err := uuid.Parse(managerIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid manager ID")
		return
	}
	
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	
	err = h.teamService.RemoveManager(teamID, userID.(uuid.UUID), managerID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	
	utils.SuccessResponse(c, http.StatusOK, "Manager removed successfully", nil)
}