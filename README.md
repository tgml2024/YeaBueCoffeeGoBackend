# YeaBueCoffeeGoBackend

ระบบ Backend สำหรับ YeaBueCoffee (Gin + GORM + MySQL) รองรับการล็อกอินด้วยประเภทผู้ใช้ Customer/Employee, จัดการ Session ลงฐานข้อมูล, และการสมัครสมาชิก (Customer/Employee แบบ Leader-only)

## ข้อกำหนดเบื้องต้น

- ติดตั้ง Go 1.21+ (หรือ 1.25 toolchain ตาม go.mod)
- ติดตั้ง MySQL และสร้างฐานข้อมูลไว้ล่วงหน้า (ตามค่าใน .env)
- แนะนำให้ติดตั้ง Postman สำหรับทดสอบ API

## เริ่มต้นใช้งาน (ตั้งแต่ Clone จนรัน)

1. Clone โปรเจค

```bash
git clone <REPO_URL>
cd YeaBueCoffeeGoBackend
```

2. สร้างไฟล์ .env ที่โฟลเดอร์ `YeaBueCoffeeGoBackend/`

```env
PORT=8112
NODE_ENV=development

DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASS=
DB_NAME=yeabue_coffee

JWT_SECRET=dev_access
REFRESH_TOKEN_SECRET=dev_refresh

COOKIE_SECURE=false
COOKIE_SAMESITE=strict
COOKIE_DOMAIN=localhost
```

3. ติดตั้ง dependency

```bash
go mod tidy
```

4. รันเซิร์ฟเวอร์

```bash
go run ./cmd/server
```

เซิร์ฟเวอร์จะเริ่มที่ `http://localhost:8112` (หรือพอร์ตที่กำหนดใน .env)

หมายเหตุ: ถ้าเรียกแล้วเจอ 404 จาก Apache แปลว่าพอร์ตชนกับ Apache บนเครื่อง ให้เปลี่ยน `PORT` ใน .env เป็นพอร์ตอื่น (เช่น 8081) แล้วรันใหม่ จากนั้นเรียก `http://localhost:8081`

## โครงสร้างหลัก

- `internal/controllers/auth_controller.go`: Handler เกี่ยวกับ login/logout/register
- `internal/middleware/*.go`: Authentication, Authorization, Token/Session อัปเดต
- `internal/models/*.go`: โครงสร้างตาราง GORM (Customer, Employee, Authen)
- `internal/db/db.go`: เชื่อมต่อ DB และ AutoMigrate
- `internal/routes/router.go`: กำหนดเส้นทาง API

## ฐานข้อมูลและการ Migrate

- ระบบใช้ GORM AutoMigrate อัตโนมัติเมื่อรันเซิร์ฟเวอร์
- ตารางที่ใช้:
  - `customers` (โมเดล `Customer`): เก็บข้อมูลลูกค้า
  - `employees` (โมเดล `Employee`): เก็บข้อมูลพนักงาน มีฟิลด์ `position` รองรับค่า "employee" หรือ "leader"
  - `authens` (โมเดล `Authen`): เก็บ session ของผู้ใช้ (`session_id`, `start_date`, `last_access`, `end_date`, `user_id`)

## การยืนยันตัวตนและ Session

- Login สำเร็จ ระบบจะออก `accessToken` (อายุ 1 ชม.) และ `refreshToken` (อายุ 30 วัน) แบบ HS256 และตั้งค่าเป็น Cookie
- ระบบสร้าง `sessionId` (สุ่ม) เก็บในตาราง `authen` พร้อมตั้ง Cookie `sessionId`
- ทุกรีเควสท์ที่ผ่าน `AuthenticateToken()` และมี `sessionId` จะอัปเดต `last_access`
- Logout จะตั้งค่า `end_date` ใน `authen` และลบคุกกี้ทั้งหมด

## ประเภทผู้ใช้

- Claims ใน JWT จะมี `userId` และ `type` (ค่าเป็น `customer` หรือ `employee`)
- การตรวจสิทธิ์:
  - `IsCustomer()` อนุญาตเฉพาะผู้ใช้ที่มี `type == "customer"`
  - `IsEmployee()` อนุญาตเฉพาะผู้ใช้ที่มี `type == "employee"`
  - `IsLeader()` ตรวจเพิ่มจาก DB ว่าผู้ใช้เป็นพนักงานที่ `position == "leader"`

## รายการ API และการใช้งาน

Base URL: `http://localhost:<PORT>` (ค่าเริ่มต้น 8112)

### 1) Login

- `POST /api/login`
  Body (JSON):

```json
{
  "username": "string",
  "password": "string",
  "type": "customer" // หรือ "employee"
}
```

ผลลัพธ์: ตั้งคุกกี้ `accessToken`, `refreshToken`, `sessionId` และตอบ JSON `{ message, type }`

### 2) Logout

- `POST /api/logout`
  ผลลัพธ์: ลบคุกกี้และปิด session (ตั้ง `end_date`)

### 3) Register Customer

- `POST /api/register/customer`
  Body (JSON):

```json
{
  "username": "cust01",
  "password": "secret",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com"
}
```

หมายเหตุ: ตรวจ username ซ้ำและแฮชรหัสผ่านก่อนบันทึก

### 4) Register Employee (Leader-only)

- `POST /api/register/employee`
- ต้องล็อกอิน และเป็น `employee` ที่มี `position == "leader"`
  Body (JSON):

```json
{
  "username": "emp02",
  "password": "secret",
  "first_name": "Jane",
  "last_name": "Smith",
  "email": "jane@example.com",
  "position": "employee" // หรือ "leader" (ถ้าไม่ส่ง จะตั้งเป็น "employee")
}
```

หมายเหตุ: ใช้ `AuthenticateToken()`, `IsEmployee()`, `IsLeader()` กับเส้นทางนี้ใน router

### 5) Customer Me (Protected)

- `GET /api/customer/me`
- ต้องล็อกอิน และผ่าน `IsCustomer()`
  ผลลัพธ์: คืน claims ที่อยู่ใน context

### 6) Employee Me (Protected)

- `GET /api/employee/me`
- ต้องล็อกอิน และผ่าน `IsEmployee()`
  ผลลัพธ์: คืน claims ที่อยู่ใน context

## ตัวอย่าง Postman Collection (ย่อ)

สามารถนำ JSON นี้ไป Import ใน Postman เพื่อทดสอบอย่างรวดเร็ว (มีในข้อความสนทนาก่อนหน้า หรือสร้างเองตาม API ข้างต้น)

## การทดสอบอย่างรวดเร็ว

1. รันเซิร์ฟเวอร์ให้ขึ้น `http://localhost:8112`
2. เรียก `POST /api/register/customer` เพื่อสร้างลูกค้าอย่างน้อย 1 ราย
3. สร้าง employee leader ด้วยขั้นตอน:
   - ก่อนอื่นคุณต้องมี leader อยู่แล้ว หรืออาจใส่โดยตรงในฐานข้อมูลชั่วคราว (ตั้ง `position=leader`)
   - ล็อกอินด้วย leader → เรียก `POST /api/register/employee` เพื่อสร้างพนักงานคนอื่นได้
4. ล็อกอินเป็น customer/employee แล้วเรียก `/api/customer/me` หรือ `/api/employee/me`

## ปัญหาพบบ่อย

- เรียก API แล้วขึ้นหน้า Apache: พอร์ตชนกับ Apache → เปลี่ยน `PORT` ใน .env เป็นพอร์ตอื่น เช่น 8081 แล้วเรียก `http://localhost:8081`
- 401/403: ตรวจว่าได้ส่งคุกกี้ (Postman จะเก็บให้อัตโนมัติ) และเส้นทางที่เรียกใช้ middleware ตรงกับประเภทผู้ใช้
- รันไม่ขึ้นเพราะ DB: ตรวจ `DB_HOST/DB_PORT/DB_USER/DB_PASS/DB_NAME` ให้เชื่อมต่อได้จริง และ user มีสิทธิ์สร้างตาราง

## หมายเหตุด้านความปลอดภัย (แนะนำปรับใช้ใน Production)

- ตั้ง `COOKIE_SECURE=true` และ `COOKIE_SAMESITE=none` เมื่อใช้ผ่าน HTTPS
- ใช้ค่า `JWT_SECRET`, `REFRESH_TOKEN_SECRET` ที่คาดเดายากและเก็บเป็นความลับ
- จำกัด CORS และตั้งค่า Firewall ตามสภาพแวดล้อมจริง
