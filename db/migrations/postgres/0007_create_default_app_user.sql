-- +goose Up
INSERT INTO users (
    id,
    first_name,
    last_name,
    status,
    email,
    experience,  -- Adding required fields from the users table
    theme,
    notification_type
) VALUES (
    '07b1d5ce-2b81-beef-9288-63b5c93ea278',
    'Rodney',
    'the Rocket Bot',
    'active',
    'rodney@rocketbox.co',
    ARRAY['other'],  -- Default value from users table
    'light',         -- Default value from users table
    'email'          -- Default value from users table
);

-- +goose Down
-- Delete in order to respect foreign key constraints
DELETE FROM organization_users 
WHERE user_id = '07b1d5ce-2b81-beef-9288-63b5c93ea278';

DELETE FROM organization_domains 
WHERE created_by_id = '07b1d5ce-2b81-beef-9288-63b5c93ea278';

UPDATE organizations 
SET created_by_id = NULL 
WHERE created_by_id = '07b1d5ce-2b81-beef-9288-63b5c93ea278';

DELETE FROM user_factors 
WHERE user_id = '07b1d5ce-2b81-beef-9288-63b5c93ea278';

DELETE FROM users 
WHERE id = '07b1d5ce-2b81-beef-9288-63b5c93ea278';