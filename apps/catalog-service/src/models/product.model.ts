import { Entity, PrimaryGeneratedColumn, Column, CreateDateColumn, UpdateDateColumn, Index } from 'typeorm';

@Entity('products')
export class Product {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({ unique: true })
  sku: string;

  @Column()
  nameEn: string;

  @Column()
  nameAr: string;

  @Column('decimal', { precision: 15, scale: 2 })
  basePrice: number;

  @Column({ default: 'SAR' })
  currency: string;

  @Column({ default: 0 })
  stockQuantity: number;

  @Column({ default: 'RIYADH_WAREHOUSE' })
  warehouseId: string;

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;
}
