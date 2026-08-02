-- AI Payment Support & Transaction Portal
-- Final SQL Schema

CREATE DATABASE IF NOT EXISTS payment_support_db;
USE payment_support_db;

CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transactions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    transaction_id VARCHAR(50) UNIQUE NOT NULL,
    user_id INT NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    status ENUM('SUCCESS','FAILED','PENDING') NOT NULL,
    failure_reason VARCHAR(255) NULL,
    payment_method ENUM('UPI','Card','Net Banking','Wallet') NOT NULL,
    receiver_name VARCHAR(100) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE refunds (
    id INT AUTO_INCREMENT PRIMARY KEY,
    transaction_id INT NOT NULL UNIQUE,
    amount NUMERIC(10,2) NOT NULL,
    reason ENUM('Duplicate Payment','Wrong Amount','Order Cancelled','Other') NOT NULL,
    status ENUM('REQUESTED','PROCESSING','COMPLETED') NOT NULL DEFAULT 'REQUESTED',
    requested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME NULL,
    FOREIGN KEY (transaction_id) REFERENCES transactions(id)
);

CREATE TABLE support_tickets (
    id INT AUTO_INCREMENT PRIMARY KEY,
    ticket_id VARCHAR(50) UNIQUE NOT NULL,
    user_id INT NOT NULL,
    transaction_id INT NOT NULL,
    subject VARCHAR(150) NOT NULL,
    description TEXT NOT NULL,
    status ENUM('OPEN','IN_PROGRESS','CLOSED') NOT NULL DEFAULT 'OPEN',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    closed_at DATETIME NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (transaction_id) REFERENCES transactions(id)
);
