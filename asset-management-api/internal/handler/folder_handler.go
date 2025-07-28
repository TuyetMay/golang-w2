package handler

import (
	"asset-management-api/internal/middleware"
	"asset-management-api/internal/service/interfaces"
	"asset-management-api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FolderHandler struct {
	folderService interfaces.FolderService
}

type CreateFolderRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

type UpdateFolderRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

func NewFolderHandler(folderService interfaces.FolderService) *FolderHandler {
	return &FolderHandler{folderService: folderService}
}

// POST /folders
func (h *FolderHandler) CreateFolder(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate request
	if errors := utils.ValidateStruct(req); len(errors) > 0 {
		utils.ValidationErrorResponse(c, utils.GetValidationErrorMessages(errors))
		return
	}

	folder, err := h.folderService.CreateFolder(userID, req.Name, req.Description)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create folder", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Folder created successfully", folder)
}

// GET /folders/:folderId
func (h *FolderHandler) GetFolder(c *gin.Context) {
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

	folder, err := h.folderService.GetFolder(folderID, userID)
	if err != nil {
		if err.Error() == "folder not found" {
			utils.NotFoundResponse(c, "Folder not found")
			return
		}
		if err.Error() == "access denied: you don't have permission to view this folder" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get folder", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Folder retrieved successfully", folder)
}

// PUT /folders/:folderId
func (h *FolderHandler) UpdateFolder(c *gin.Context) {
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

	var req UpdateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate request
	if errors := utils.ValidateStruct(req); len(errors) > 0 {
		utils.ValidationErrorResponse(c, utils.GetValidationErrorMessages(errors))
		return
	}

	folder, err := h.folderService.UpdateFolder(folderID, userID, req.Name, req.Description)
	if err != nil {
		if err.Error() == "folder not found" {
			utils.NotFoundResponse(c, "Folder not found")
			return
		}
		if err.Error() == "access denied: you don't have write permission for this folder" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update folder", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Folder updated successfully", folder)
}

// DELETE /folders/:folderId
func (h *FolderHandler) DeleteFolder(c *gin.Context) {
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

	err = h.folderService.DeleteFolder(folderID, userID)
	if err != nil {
		if err.Error() == "access denied: only the folder owner can delete it" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to delete folder", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Folder deleted successfully", nil)
}

// GET /folders (Get user's folders)
func (h *FolderHandler) GetUserFolders(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	folders, err := h.folderService.GetUserFolders(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get folders", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Folders retrieved successfully", folders)
}