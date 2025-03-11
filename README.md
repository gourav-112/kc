# Kirana Club

## Overview
Kirana Club is a job-processing system designed to handle multiple store visits within a single job. Each store visit contains multiple images, and the job is considered **completed** only when all images are processed successfully. If any image processing fails, the job is marked as **failed**. The system uses **Golang** for backend development and **GORM** as an ORM to interact with an SQLite database.

## Features
- Submit a job containing multiple store visits and images.
- Track the status of a job.
- Concurrent image processing with a success/failure mechanism.
- Uses **GORM** for easy database management.

## Project Structure
```
/kc
│── /dbs                # Database initialization and models
│── /helper             # Utility functions
│── /service            # Business logic for job processing
│── main.go             # Entry point of the application
│── go.mod              # Go module dependencies
│── go.sum              # Dependency checksums
│── README.md           # Project documentation
```

## Technologies Used
- **Golang** (Backend)
- **SQLite** (Database)
- **net/http** (API development)
- **sync** (Concurrency handling)
- **GORM** and **GIN** can be used further to optimize SQL queries.

---

## Setup Instructions

### Prerequisites
Make sure you have **Go** installed. If not, download it from [here](https://golang.org/dl/).

### Clone the Repository
```sh
git clone https://github.com/gourav-112/kc.git
cd kc
```

### Install Dependencies
```sh
go mod tidy
```

### Run the Application
```sh
go run main.go
```
The server will start at `http://localhost:8080/`

---

## Database Schema
The system uses **GORM** for database interactions. The tables are structured as follows:

### **Jobs Table**
| Column  | Type   | Description                          |
|---------|--------|--------------------------------------|
| id      | INT (PK) | Unique Job ID                     |
| status  | STRING  | Job status (ongoing/completed/failed) |

### **Visits Table**
| Column  | Type   | Description                          |
|---------|--------|--------------------------------------|
| id      | INT (PK) | Unique Visit ID                   |
| job_id  | INT (FK) | Associated Job ID                 |
| store_id | STRING | Store Identifier                   |

### **Images Table**
| Column  | Type   | Description                          |
|---------|--------|--------------------------------------|
| id      | INT (PK) | Unique Image ID                   |
| visit_id | INT (FK) | Associated Visit ID              |
| image_url | STRING | Image URL                         |
| status  | STRING  | Image status (ongoing/completed/failed) |

---

## API Endpoints

### **1. Submit a Job**
**Endpoint:** `POST /api/submit`

**Request Body:**
```json
{
  "count": 2,
  "visits": [
    {
      "store_id": "S00339218",
      "image_url": [
        "https://example.com/image1.jpg",
        "https://example.com/image2.jpg"
      ],
      "visit_time": "2025-03-11T12:00:00Z"
    },
    {
      "store_id": "S01408764",
      "image_url": [
        "https://example.com/image3.jpg"
      ],
      "visit_time": "2025-03-11T12:30:00Z"
    }
  ]
}
```

**Response:**
```json
{
  "job_id": 123
}
```

### **2. Get Job Status**
**Endpoint:** `GET /api/status?jobid=123`

**Response:**
```json
{
  "job_id": 123,
  "status": "completed"
}
```

---

## How Job Processing Works
1. The **submit job** API takes a payload with multiple store visits and images.
2. The backend **creates a job** entry and associates visits and images with it.
3. The system **processes images concurrently**.
4. If all images are processed successfully, the **job is marked as completed**.
5. If any image fails, the **job is marked as failed**.
6. The **status API** helps in tracking the job status.

---

## Contribution Guide
1. Fork the repository.
2. Create a new branch: `git checkout -b feature-branch`
3. Commit your changes: `git commit -m "Added a new feature"`
4. Push to your branch: `git push origin feature-branch`
5. Submit a pull request.

---

## License
This project is open-source under the **MIT License**.

## Contact
For any queries or issues, feel free to reach out!

📧 Email: gouravbeniwal1746@gmail.com  
📌 GitHub Issues: [Report an issue](https://github.com/gourav-112/kc/issues)
