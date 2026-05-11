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
