import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository, Like } from 'typeorm';
import { Product } from '../models/product.model';

@Injectable()
export class CatalogService {
  constructor(
    @InjectRepository(Product)
    private readonly productRepo: Repository<Product>,
  ) {}

  async createProduct(dto: any, correlationId?: string): Promise<Product> {
    const product = this.productRepo.create({
      sku: dto.sku,
      nameEn: dto.nameEn,
      nameAr: dto.nameAr,
      basePrice: dto.basePrice,
      currency: dto.currency || 'SAR',
      stockQuantity: dto.stockQuantity || 0,
      warehouseId: dto.warehouseId || 'RIYADH_WAREHOUSE',
    });

    const saved = await this.productRepo.save(product);

    // TODO: Publish ProductCreated event to Kafka
    console.log(`[CATALOG] Product created id=${saved.id} correlation=${correlationId || 'N/A'}`);

    return saved;
  }

  async findById(id: string): Promise<Product> {
    const product = await this.productRepo.findOne({ where: { id } });
    if (!product) {
      throw new NotFoundException(`Product ${id} not found`);
    }
    return product;
  }

  async listProducts(
    query?: string,
    warehouseId?: string,
    page = 1,
    limit = 20,
  ): Promise<{ data: Product[]; total: number; page: number; limit: number }> {
    const where: any = {};

    if (warehouseId) {
      where.warehouseId = warehouseId;
    }

    if (query) {
      where.nameEn = Like(`%${query}%`);
    }

    const [data, total] = await this.productRepo.findAndCount({
      where,
      skip: (page - 1) * limit,
      take: limit,
      order: { createdAt: 'DESC' },
    });

    return { data, total, page, limit };
  }
}
