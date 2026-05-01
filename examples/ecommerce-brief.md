# Simple Ecommerce App

A minimal online store where users can browse products, add them to a cart, and check out.

## What's in it

### Pages
- **Home** — hero banner + featured products grid
- **Products** — browsable catalog with category filter and search
- **Product Detail** — image, title, price, description, add-to-cart button
- **Cart** — list of items, quantities, subtotal, checkout button
- **Checkout** — shipping address form + payment (card fields) + order summary
- **Order Confirmation** — success message with order ID

### Features
- User signup / login (email + password)
- Session persistence (stay logged in on refresh)
- Product catalog (name, image, price, stock count, category)
- Cart stored per user, persisted across sessions
- Basic search by product name
- Filter products by category
- Stripe payment (test mode)
- Order history visible in user account

### Data
- **Users** — id, email, password hash, created_at
- **Products** — id, name, description, price, image_url, category, stock
- **Cart Items** — user_id, product_id, quantity
- **Orders** — id, user_id, total, status, created_at
- **Order Items** — order_id, product_id, quantity, unit_price

### Tech
- Frontend: React (Vite), Tailwind CSS
- Backend: REST API (Go or Bun/TypeScript)
- Database: PostgreSQL
- Auth: JWT tokens
- Payments: Stripe Checkout
- File storage: local or S3 for product images

### Out of scope (for now)
- Admin dashboard
- Reviews / ratings
- Discount codes
- Email notifications
- Multi-currency
