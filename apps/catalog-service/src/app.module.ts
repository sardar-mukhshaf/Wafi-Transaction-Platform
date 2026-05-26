import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { CatalogController } from './controllers/catalog.controller';
import { CatalogService } from './services/catalog.service';
import { Product } from './models/product.model';

@Module({
  imports: [
    TypeOrmModule.forRoot({
      type: 'postgres',
      host: process.env.DB_HOST || 'localhost',
      port: parseInt(process.env.DB_PORT || '5432', 10),
      username: process.env.DB_USER || 'catalog_user',
      password: process.env.DB_PASSWORD || 'catalog_pass',
      database: process.env.DB_NAME || 'catalog_db',
      entities: [Product],
      synchronize: process.env.APP_ENV === 'development',
      logging: false,
    }),
    TypeOrmModule.forFeature([Product]),
  ],
  controllers: [CatalogController],
  providers: [CatalogService],
})
export class AppModule {}
