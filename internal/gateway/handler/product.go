package handler

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"net/http"
	"strconv"
)

type ProductHandler struct {
	productService service.ProductService
	uploadService  service.UploadService
}

// CreateProduct docs
// @Summary Create a new product
// @Description Create a new product (Admin only)
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateProductRequest true "Product data"
// @Success 201 {object} helper.Response{data=dto.ProductResponse} "Product created successfully"
// @Failure 400 {object} helper.Response "Invalid request data"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Failure 403 {object} helper.Response "Admin access required"
// @Router /products [post]
func (p *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreateProductRequest
	if err := helper.ReadJSON(w, r, &payload); err != nil {
		helper.BadRequestResponse(w, "Invalid given payload", err)
		return
	}

	v := helper.NewValidator()
	dto.ValidateCreateProductRequest(v, &payload)
	if !v.Valid() {
		helper.FailedValidationResponse(w, "payload is not valid")
		return
	}

	product, err := p.productService.CreateProduct(r.Context(), &payload)
	if err != nil {
		helper.InternalServerError(w, "failed to create the product", err)
		return
	}

	helper.CreatedResponse(w, "product successfully created", product)
}

// GetProducts
// @Summary Get all products
// @Description Retrieve paginated list of active products
// @Tags Products
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} helper.PaginatedResponse{data=[]dto.ProductResponse} "Products retrieved successfully"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /products [get]
func (p *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	products, meta, err := p.productService.GetProducts(r.Context(), page, limit)
	if err != nil {
		helper.InternalServerError(w, "", err)
		return
	}

	helper.PaginatedSuccessResponse(w, "Products successfully retrieved", products, *meta)
}

// GetProduct docs
// @Summary Get a product by ID
// @Description Retrieve detailed information about a specific product
// @Tags Products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} helper.Response{data=dto.ProductResponse} "Product retrieved successfully"
// @Failure 400 {object} helper.Response "Invalid product ID"
// @Failure 404 {object} helper.Response "Product not found"
// @Router /products/{id} [get]
func (p *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := helper.ReadParams(r)
	if err != nil {
		helper.BadRequestResponse(w, "Invalid id", err)
		return
	}

	product, err := p.productService.GetProductById(r.Context(), id)
	if err != nil {
		helper.InternalServerError(w, "failed to get the product", err)
		return
	}

	helper.SuccessResponse(w, "product successfully retrieved", product)
}

// UpdateProduct docs
// @Summary Update a product
// @Description Update an existing product (Admin only)
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param request body dto.UpdateProductRequest true "Product update data"
// @Success 200 {object} helper.Response{data=dto.ProductResponse} "Product updated successfully"
// @Failure 400 {object} helper.Response "Invalid request data"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Failure 403 {object} helper.Response "Admin access required"
// @Router /products/{id} [put]
func (p *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := helper.ReadParams(r)
	if err != nil {
		helper.BadRequestResponse(w, "Invalid id", err)
		return
	}

	var payload dto.UpdateProductRequest
	if err := helper.ReadJSON(w, r, &payload); err != nil {
		helper.BadRequestResponse(w, "Invalid payload", err)
		return
	}

	v := helper.NewValidator()
	dto.ValidateUpdateProductRequest(v, &payload)
	if !v.Valid() {
		helper.FailedValidationResponse(w, "payload is not valid")
		return
	}

	product, err := p.productService.UpdateProduct(r.Context(), id, &payload)
	if err != nil {
		helper.InternalServerError(w, "failed to update the product", err)
		return
	}

	helper.SuccessResponse(w, "product successfully updated", product)
}

// DeleteProduct docs
// @Summary Delete a product
// @Description Delete a product (Admin only)
// @Tags Products
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 200 {object} helper.Response "Product deleted successfully"
// @Failure 400 {object} helper.Response "Invalid product ID"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Failure 403 {object} helper.Response "Admin access required"
// @Router /products/{id} [delete]
func (p *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := helper.ReadParams(r)
	if err != nil {
		helper.BadRequestResponse(w, "Invalid id", err)
		return
	}

	if err := p.productService.DeleteProduct(r.Context(), id); err != nil {
		helper.InternalServerError(w, "failed to delete the product", err)
		return
	}

	helper.CreatedResponse(w, "product successfully deleted", nil)
}

// UploadProductImage docs
// @Summary Upload product image
// @Description Upload an image for a product (Admin only)
// @Tags Products
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param image formData file true "Image file"
// @Success 200 {object} helper.Response{data=map[string]string} "Image uploaded successfully"
// @Failure 400 {object} helper.Response "Invalid request or file"
// @Failure 401 {object} helper.Response "Unauthorized"
// @Failure 403 {object} helper.Response "Admin access required"
// @Router /products/{id}/image [post]
func (p *ProductHandler) UploadProductImage(w http.ResponseWriter, r *http.Request) {
	id, err := helper.ReadParams(r)
	if err != nil {
		helper.BadRequestResponse(w, "Invalid id", err)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		helper.BadRequestResponse(w, "No file uploaded", err)
		return
	}

	defer file.Close()

	url, err := p.uploadService.UploadProductImage(id, header)
	if err != nil {
		helper.InternalServerError(w, "failed to upload product image", err)
		return
	}

	if err := p.productService.AddProductImage(r.Context(), id, url, header.Filename); err != nil {
		helper.InternalServerError(w, "failed to upload product image", err)
		return
	}

	helper.CreatedResponse(w, "product image successfully uploaded", url)
}

func NewProductHandler(productService service.ProductService, uploadService service.UploadService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		uploadService:  uploadService,
	}
}
