

## Backend integration

The frontend is wired to the backend API exactly under `/api/v1`:

- `POST /register`
- `POST /login`
- `POST /refresh`
- `POST /logout`
- `GET /fillings`
- `GET /fillings/{fillingId}`
- `POST /fillings` (admin)
- `PATCH /fillings/{fillingId}` (admin)
- `DELETE /fillings/{fillingId}` (admin)
- `POST /orders/price`
- `POST /orders`
- `GET /orders/my`
- `GET /orders/{orderId}`
- `GET /orders` (admin)
- `PATCH /orders/{orderId}/status` (admin)

The backend directory in the integration archive is copied from the supplied backend archive and was not modified.

Run frontend with `npm run dev`; Vite proxies `/api` to `http://localhost:8080`.

Important: the supplied backend currently contains `ErrNotImplemented` stubs in fillings and orders services/handlers, so those endpoints are wired in the frontend but cannot return real fillings/orders until the backend implementation is completed. The frontend does not alter or replace those backend files.
