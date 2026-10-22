-- Sample test data for AI Payment Support & Transaction Portal
-- Run AFTER database_schema.sql
-- Note: password_hash values below are placeholders. Register real
-- users through the API (or set ADMIN_EMAIL / ADMIN_PASSWORD) so
-- valid password hashes are generated.

USE payment_support_db;

INSERT INTO users (name, email, password_hash, role) VALUES
('Ugesh Kumar', 'ugesh@example.com', '$2b$12$examplehashvalue1', 'user'),
('Admin User', 'admin@example.com', '$2b$12$examplehashvalue2', 'admin');

INSERT INTO transactions (transaction_id, user_id, amount, status, failure_reason, payment_method, receiver_name) VALUES
('TXN-9F2A4C8B', 1, 999.00, 'SUCCESS', NULL, 'UPI', 'Amazon India'),
('TXN-AB1234CD', 1, 450.00, 'FAILED', 'Insufficient Balance', 'Card', 'Flipkart'),
('TXN-77XYZ001', 1, 1200.00, 'PENDING', NULL, 'Net Banking', 'Swiggy'),
('TXN-8899QQRR', 1, 5000.00, 'SUCCESS', NULL, 'Wallet', 'Myntra');

INSERT INTO refunds (transaction_id, amount, reason, status) VALUES
(1, 999.00, 'Order Cancelled', 'PROCESSING');

INSERT INTO support_tickets (ticket_id, user_id, transaction_id, subject, description, status) VALUES
('TKT-4C2A9F1B', 1, 2, 'Payment failed but amount deducted', 'My card shows a deduction but the transaction failed on your end.', 'OPEN');
