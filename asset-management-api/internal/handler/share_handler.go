package handler

import (
	"asset-management-api/internal/middleware"
	"asset-management-api/internal/models"
	"asset-management-api/internal/service/interfaces"
	"asset-management-api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ShareHandler struct {
	shareService interfaces.ShareService
}

func NewShareHandler(shareService interfaces.ShareService) *ShareHandler {
	return &ShareHandler{shareService: shareService}
}

// POST /folders/:folderId/share
func (h *ShareHandler) ShareFolder(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	folderIDStr := c.Param("folderId")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid folder ID format", err)
		return
	}

	var req models.ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate request
	if errors := utils.ValidateStruct(req); len(errors) > 0 {
		utils.ValidationErrorResponse(c, utils.GetValidationErrorMessages(errors))
		return
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	err = h.shareService.ShareFolder(folderID, userID, targetUserID, req.AccessLevel)
	if err != nil {
		if err.Error() == "access denied: only the folder owner can share it" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		if err.Error() == "cannot share folder with yourself" {
			utils.BadRequestResponse(c, "Cannot share with yourself", err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to share folder", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Folder shared successfully", nil)
}

// DELETE /folders/:folderId/share/:userId
func (h *ShareHandler) UnshareFolder(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	folderIDStr := c.Param("folderId")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid folder ID format", err)
		return
	}

	targetUserIDStr := c.Param("userId")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	err = h.shareService.UnshareFolder(folderID, userID, targetUserID)
	if err != nil {
		if err.Error() == "access denied: only the folder owner can unshare it" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to unshare folder", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Folder unshared successfully", nil)
}

// GET /folders/:folderId/shares
func (h *ShareHandler) GetFolderShares(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	folderIDStr := c.Param("folderId")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid folder ID format", err)
		return
	}

	shares, err := h.shareService.GetFolderShares(folderID, userID)
	if err != nil {
		if err.Error() == "access denied: only the folder owner can view shares" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get folder shares", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Folder shares retrieved successfully", shares)
}

// POST /notes/:noteId/share
func (h *ShareHandler) ShareNote(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	noteIDStr := c.Param("noteId")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid note ID format", err)
		return
	}

	var req models.ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate request
	if errors := utils.ValidateStruct(req); len(errors) > 0 {
		utils.ValidationErrorResponse(c, utils.GetValidationErrorMessages(errors))
		return
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	err = h.shareService.ShareNote(noteID, userID, targetUserID, req.AccessLevel)
	if err != nil {
		if err.Error() == "access denied: only the note owner can share it" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		if err.Error() == "cannot share note with yourself" {
			utils.BadRequestResponse(c, "Cannot share with yourself", err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to share note", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Note shared successfully", nil)
}

// DELETE /notes/:noteId/share/:userId
func (h *ShareHandler) UnshareNote(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	noteIDStr := c.Param("noteId")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid note ID format", err)
		return
	}

	targetUserIDStr := c.Param("userId")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	err = h.shareService.UnshareNote(noteID, userID, targetUserID)
	if err != nil {
		if err.Error() == "access denied: only the note owner can unshare it" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to unshare note", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Note unshared successfully", nil)
}

// GET /notes/:noteId/shares
func (h *ShareHandler) GetNoteShares(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	noteIDStr := c.Param("noteId")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid note ID format", err)
		return
	}

	shares, err := h.shareService.GetNoteShares(noteID, userID)
	if err != nil {
		if err.Error() == "access denied: only the note owner can view shares" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get note shares", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Note shares retrieved successfully", shares)
}