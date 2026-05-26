import { Controller, Get, Post, Body, Param, Query, Headers } from '@nestjs/common';
import { ApiTags, ApiOperation, ApiResponse } from '@nestjs/swagger';
import { CatalogService } from '../services/catalog.service';
import { Product } from '../models/product.model';

class CreateProductDto {
  sku: string;
  nameEn: string;
  nameAr: string;
  basePrice: number;
  currency?: string;
  stockQuantity?: number;
  warehouseId?: string;
}

@ApiTags('products')
@Controller('products')
export class CatalogController {
  constructor(private readonly catalogService: CatalogService) {}

  @Post()
  @ApiOperation({ summary: 'Create a new product' })
  @ApiResponse({ status: 201, description: 'Product created', type: Product })
  async createProduct(
    @Body() dto: CreateProductDto,
    @Headers('x-correlation-id') correlationId?: string,
  ): Promise<Product> {
    return this.catalogService.createProduct(dto, correlationId);
  }

  @Get(':id')
  @ApiOperation({ summary: 'Get product by ID' })
  @ApiResponse({ status: 200, description: 'Product found', type: Product })
  @ApiResponse({ status: 404, description: 'Product not found' })
  async getProduct(@Param('id') id: string): Promise<Product> {
    return this.catalogService.findById(id);
  }

  @Get()
  @ApiOperation({ summary: 'List products with optional search' })
  @ApiResponse({ status: 200, description: 'List of products', type: [Product] })
  async listProducts(
    @Query('q') query?: string,
    @Query('warehouse') warehouseId?: string,
    @Query('page') page = 1,
    @Query('limit') limit = 20,
  ): Promise<{ data: Product[]; total: number; page: number; limit: number }> {
    return this.catalogService.listProducts(query, warehouseId, page, limit);
  }

  @Get('health')
  health() {
    return { status: 'healthy', service: 'catalog-service' };
  }
}
