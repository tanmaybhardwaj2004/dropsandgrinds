-- Seed data for Indian payment offers
-- This includes UPI discounts, card cashback, and wallet bonuses

-- UPI Payment Offers
INSERT INTO indian_payment_offers (store_id, offer_type, provider, description, discount_percent, max_discount_amount, min_order_amount, valid_from, valid_until, is_active, created_at) VALUES
-- Steam UPI offers
(1, 'upi_discount', 'phonepe', '10% instant discount on Steam purchases using PhonePe UPI', 10, 200.00, 500.00, NOW(), NOW() + INTERVAL '90 days', TRUE, NOW()),
(1, 'upi_discount', 'gpay', '10% cashback on Steam games using Google Pay UPI', 10, 150.00, 500.00, NOW(), NOW() + INTERVAL '90 days', TRUE, NOW()),
(1, 'upi_discount', 'paytm', '5% discount on Steam via Paytm UPI', 5, 100.00, 299.00, NOW(), NOW() + INTERVAL '60 days', TRUE, NOW()),

-- Epic Games UPI offers
(2, 'upi_discount', 'phonepe', '15% discount on Epic Games store with PhonePe', 15, 300.00, 500.00, NOW(), NOW() + INTERVAL '90 days', TRUE, NOW()),
(2, 'upi_discount', 'gpay', '12% cashback on Epic Games via Google Pay', 12, 250.00, 500.00, NOW(), NOW() + INTERVAL '90 days', TRUE, NOW()),

-- GreenManGaming UPI offers
(6, 'upi_discount', 'paytm', '8% discount on GreenManGaming with Paytm UPI', 8, 120.00, 399.00, NOW(), NOW() + INTERVAL '60 days', TRUE, NOW()),

-- Fanatical UPI offers
(7, 'upi_discount', 'phonepe', '10% off Fanatical bundles via PhonePe', 10, 200.00, 499.00, NOW(), NOW() + INTERVAL '75 days', TRUE, NOW()),

-- Humble Bundle UPI offers
(8, 'upi_discount', 'gpay', '5% discount on Humble Bundle with Google Pay', 5, 100.00, 199.00, NOW(), NOW() + INTERVAL '60 days', TRUE, NOW());

-- Card Payment Offers
INSERT INTO indian_payment_offers (store_id, offer_type, provider, description, discount_percent, max_discount_amount, min_order_amount, valid_from, valid_until, is_active, created_at) VALUES
-- Steam card offers
(1, 'card_discount', 'hdfc', '10% discount on Steam using HDFC debit/credit cards', 10, 500.00, 999.00, NOW(), NOW() + INTERVAL '120 days', TRUE, NOW()),
(1, 'card_discount', 'icici', '10% off Steam games with ICICI cards', 10, 500.00, 999.00, NOW(), NOW() + INTERVAL '120 days', TRUE, NOW()),
(1, 'card_discount', 'sbi', '5% discount on Steam via SBI cards', 5, 300.00, 499.00, NOW(), NOW() + INTERVAL '90 days', TRUE, NOW()),
(1, 'card_discount', 'axis', '8% off Steam with Axis Bank cards', 8, 400.00, 799.00, NOW(), NOW() + INTERVAL '90 days', TRUE, NOW()),

-- Epic Games card offers
(2, 'card_discount', 'hdfc', '15% discount on Epic Games with HDFC cards', 15, 600.00, 999.00, NOW(), NOW() + INTERVAL '120 days', TRUE, NOW()),
(2, 'card_discount', 'icici', '12% off Epic Games using ICICI cards', 12, 500.00, 999.00, NOW(), NOW() + INTERVAL '120 days', TRUE, NOW()),

-- Xbox card offers
(3, 'card_discount', 'hdfc', '10% discount on Xbox Game Pass with HDFC cards', 10, 400.00, 799.00, NOW(), NOW() + INTERVAL '90 days', TRUE, NOW()),
(3, 'card_discount', 'axis', '8% off Xbox games with Axis cards', 8, 350.00, 699.00, NOW(), NOW() + INTERVAL '90 days', TRUE, NOW()),

-- PlayStation card offers
(4, 'card_discount', 'icici', '10% discount on PlayStation Store with ICICI cards', 10, 500.00, 999.00, NOW(), NOW() + INTERVAL '90 days', TRUE, NOW()),
(4, 'card_discount', 'sbi', '8% off PlayStation games via SBI cards', 8, 400.00, 799.00, NOW(), NOW() + INTERVAL '90 days', TRUE, NOW());

-- Wallet Bonus Offers
INSERT INTO indian_payment_offers (store_id, offer_type, provider, description, discount_percent, max_discount_amount, min_order_amount, valid_from, valid_until, is_active, created_at) VALUES
-- Paytm wallet offers
(1, 'wallet_bonus', 'paytm', 'Get 5% Paytm cashback on Steam purchases via Paytm wallet', 5, 200.00, 399.00, NOW(), NOW() + INTERVAL '60 days', TRUE, NOW()),
(2, 'wallet_bonus', 'paytm', '8% Paytm cashback on Epic Games via wallet', 8, 250.00, 499.00, NOW(), NOW() + INTERVAL '60 days', TRUE, NOW()),
(6, 'wallet_bonus', 'paytm', '6% cashback on GreenManGaming via Paytm wallet', 6, 150.00, 399.00, NOW(), NOW() + INTERVAL '60 days', TRUE, NOW()),

-- Amazon Pay wallet offers
(1, 'wallet_bonus', 'amazonpay', '10% Amazon Pay cashback on Steam', 10, 300.00, 599.00, NOW(), NOW() + INTERVAL '75 days', TRUE, NOW()),
(2, 'wallet_bonus', 'amazonpay', '12% Amazon Pay cashback on Epic Games', 12, 400.00, 699.00, NOW(), NOW() + INTERVAL '75 days', TRUE, NOW()),
(3, 'wallet_bonus', 'amazonpay', '8% Amazon Pay cashback on Xbox', 8, 300.00, 599.00, NOW(), NOW() + INTERVAL '75 days', TRUE, NOW()),

-- Freecharge wallet offers
(1, 'wallet_bonus', 'freecharge', '5% Freecharge cashback on Steam purchases', 5, 150.00, 299.00, NOW(), NOW() + INTERVAL '60 days', TRUE, NOW()),
(6, 'wallet_bonus', 'freecharge', '7% Freecharge cashback on GreenManGaming', 7, 180.00, 399.00, NOW(), NOW() + INTERVAL '60 days', TRUE, NOW());

-- Special Bundle Offers
INSERT INTO indian_payment_offers (store_id, offer_type, provider, description, discount_percent, max_discount_amount, min_order_amount, valid_from, valid_until, is_active, created_at) VALUES
-- Steam bundle offers
(1, 'bundle_discount', 'phonepe', '15% extra discount on Steam bundles with PhonePe UPI', 15, 500.00, 999.00, NOW(), NOW() + INTERVAL '90 days', TRUE, NOW()),
(1, 'bundle_discount', 'hdfc', '20% discount on Steam bundles with HDFC cards', 20, 800.00, 1499.00, NOW(), NOW() + INTERVAL '120 days', TRUE, NOW()),

-- Epic Games bundle offers
(2, 'bundle_discount', 'gpay', '18% extra discount on Epic bundles with Google Pay', 18, 600.00, 1199.00, NOW(), NOW() + INTERVAL '90 days', TRUE, NOW()),
(2, 'bundle_discount', 'icici', '20% off Epic bundles with ICICI cards', 20, 700.00, 1299.00, NOW(), NOW() + INTERVAL '120 days', TRUE, NOW()),

-- Fanatical bundle offers
(7, 'bundle_discount', 'phonepe', '12% extra discount on Fanatical bundles with PhonePe', 12, 400.00, 899.00, NOW(), NOW() + INTERVAL '75 days', TRUE, NOW()),
(7, 'bundle_discount', 'paytm', '10% off Fanatical bundles via Paytm wallet', 10, 350.00, 799.00, NOW(), NOW() + INTERVAL '75 days', TRUE, NOW());
