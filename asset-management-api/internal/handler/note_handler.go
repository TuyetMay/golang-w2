package handler

import (
	"asset-management-api/internal/middleware"
	"asset-management-api/internal/service/interfaces"
	"asset-management-api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NoteHandler struct {
	noteService interfaces.NoteService
}

type CreateNoteRequest struct {
	Title string `json:"title" validate:"required,min=1,max=255"`
	Body  string `json:"body" validate:"max=10000"`
}

type UpdateNoteRequest struct {
	Title string `json:"title" validate:"required,min=1,max=255"`
	Body  string `json:"body" validate:"max=10000"`
}

func NewNoteHandler(noteService interfaces.NoteService) *NoteHandler {
	return &NoteHandler{noteService: noteService}
}

// POST /folders/:folderId/notes
func (h *NoteHandler) CreateNote(c *gin.Context) {
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

	var req CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate request
	if errors := utils.ValidateStruct(req); len(errors) > 0 {
		utils.ValidationErrorResponse(c, utils.GetValidationErrorMessages(errors))
		return
	}

	note, err := h.noteService.CreateNote(userID, folderID, req.Title, req.Body)
	if err != nil {
		if err.Error() == "access denied: you don't have write permission for this folder" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to create note", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Note created successfully", note)
}

// GET /notes/:noteId
func (h *NoteHandler) GetNote(c *gin.Context) {
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

	note, err := h.noteService.GetNote(noteID, userID)
	if err != nil {
		if err.Error() == "note not found" {
			utils.NotFoundResponse(c, "Note not found")
			return
		}
		if err.Error() == "access denied: you don't have permission to view this note" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get note", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Note retrieved successfully", note)
}

// PUT /notes/:noteId
func (h *NoteHandler) UpdateNote(c *gin.Context) {
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

	var req UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate request
	if errors := utils.ValidateStruct(req); len(errors) > 0 {
		utils.ValidationErrorResponse(c, utils.GetValidationErrorMessages(errors))
		return
	}

	note, err := h.noteService.UpdateNote(noteID, userID, req.Title, req.Body)
	if err != nil {
		if err.Error() == "note not found" {
			utils.NotFoundResponse(c, "Note not found")
			return
		}
		if err.Error() == "access denied: you don't have write permission for this note" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update note", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Note updated successfully", note)
}

// DELETE /notes/:noteId
func (h *NoteHandler) DeleteNote(c *gin.Context) {
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

	err = h.noteService.DeleteNote(noteID, userID)
	if err != nil {
		if err.Error() == "access denied: only the note owner can delete it" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to delete note", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Note deleted successfully", nil)
}

// GET /folders/:folderId/notes
func (h *NoteHandler) GetNotesByFolder(c *gin.Context) {
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

	notes, err := h.noteService.GetNotesByFolder(folderID, userID)
	if err != nil {
		if err.Error() == "access denied: you don't have permission to view this folder" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get notes", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notes retrieved successfully", notes)
}

// GET /notes (Get user's notes)
func (h *NoteHandler) GetUserNotes(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	notes, err := h.noteService.GetUserNotes(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get notes", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notes retrieved successfully", notes)
}