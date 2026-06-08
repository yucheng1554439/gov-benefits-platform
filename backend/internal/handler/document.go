package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/govbenefits/platform/internal/middleware"
	"github.com/govbenefits/platform/internal/service"
)

type DocumentHandler struct {
	docs *service.DocumentService
}

func NewDocumentHandler(docs *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{docs: docs}
}

func (h *DocumentHandler) List(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	docs, err := h.docs.List(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": docs})
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	doc, err := h.docs.Upload(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), caseID, header.Filename, mimeType, header.Size, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, doc)
}

func (h *DocumentHandler) Download(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}
	reader, doc, err := h.docs.Download(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), docID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", "attachment; filename="+doc.OriginalName)
	c.Header("Content-Type", doc.MimeType)
	c.Header("Content-Length", strconv.FormatInt(doc.FileSize, 10))
	_, _ = io.Copy(c.Writer, reader)
}

type verifyRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *DocumentHandler) Verify(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}
	var req verifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.docs.Verify(c.Request.Context(), middleware.GetAgencyID(c), middleware.GetUserID(c), docID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "verified"})
}
