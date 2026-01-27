-- Migration: Agent Profile Field Metadata Table
-- Phase 6.2: Dynamic Field Metadata for Update Form
-- Stores metadata about profile fields: editability, validation, approval requirements, display info

-- Table: agent_profile_field_metadata
CREATE TABLE IF NOT EXISTS agent_profile_field_metadata (
    field_id SERIAL PRIMARY KEY,

    -- Field identification
    field_name TEXT NOT NULL UNIQUE,
    section TEXT NOT NULL CHECK (section IN ('personal_info', 'address', 'contact', 'email', 'bank', 'license')),

    -- Display information
    display_name TEXT NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    field_type TEXT NOT NULL CHECK (field_type IN ('text', 'number', 'date', 'email', 'phone', 'select', 'textarea', 'checkbox')),
    placeholder TEXT,
    help_text TEXT,

    -- Editability and validation
    is_editable BOOLEAN NOT NULL DEFAULT true,
    is_required BOOLEAN NOT NULL DEFAULT false,
    requires_approval BOOLEAN NOT NULL DEFAULT false,

    -- Validation rules (JSONB for flexibility)
    validation_rules JSONB,
    -- Example: {"min_length": 3, "max_length": 100, "pattern": "^[A-Z0-9]+$"}

    -- Select options (for dropdown fields)
    select_options JSONB,
    -- Example: [{"value": "MALE", "label": "Male"}, {"value": "FEMALE", "label": "Female"}]

    -- Metadata
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_field_metadata_section ON agent_profile_field_metadata(section) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_field_metadata_editable ON agent_profile_field_metadata(is_editable) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_field_metadata_approval ON agent_profile_field_metadata(requires_approval) WHERE is_active = true;

-- Seed data for personal_info section
INSERT INTO agent_profile_field_metadata (
    field_name, section, display_name, display_order, field_type,
    is_editable, is_required, requires_approval, validation_rules, placeholder, help_text
) VALUES
-- Critical fields (require approval)
('first_name', 'personal_info', 'First Name', 1, 'text', true, true, true,
    '{"min_length": 2, "max_length": 50, "pattern": "^[A-Za-z ]+$"}',
    'Enter first name', 'Legal first name as per documents'),
('middle_name', 'personal_info', 'Middle Name', 2, 'text', true, false, true,
    '{"max_length": 50, "pattern": "^[A-Za-z ]*$"}',
    'Enter middle name (optional)', 'Middle name if applicable'),
('last_name', 'personal_info', 'Last Name', 3, 'text', true, true, true,
    '{"min_length": 2, "max_length": 50, "pattern": "^[A-Za-z ]+$"}',
    'Enter last name', 'Legal last name as per documents'),
('pan_number', 'personal_info', 'PAN Number', 4, 'text', true, true, true,
    '{"length": 10, "pattern": "^[A-Z]{5}[0-9]{4}[A-Z]{1}$"}',
    'AAAAA0000A', 'Permanent Account Number (PAN) - 10 characters'),
('aadhar_number', 'personal_info', 'Aadhar Number', 5, 'text', true, true, true,
    '{"length": 12, "pattern": "^[0-9]{12}$"}',
    '000000000000', 'Aadhar Number - 12 digits'),

-- Non-critical fields (immediate update)
('title', 'personal_info', 'Title', 0, 'select', true, false, false,
    '{}', 'Select title', 'Mr./Mrs./Ms./Dr.'),
('gender', 'personal_info', 'Gender', 6, 'select', true, true, false,
    '{}', 'Select gender', 'Gender as per documents'),
('date_of_birth', 'personal_info', 'Date of Birth', 7, 'date', true, true, false,
    '{"min": "1950-01-01", "max": "2010-12-31"}',
    'YYYY-MM-DD', 'Date of birth'),
('marital_status', 'personal_info', 'Marital Status', 8, 'select', true, false, false,
    '{}', 'Select marital status', 'Current marital status'),
('category', 'personal_info', 'Category', 9, 'select', true, false, false,
    '{}', 'Select category', 'Category (General/OBC/SC/ST)'),
('professional_title', 'personal_info', 'Professional Title', 10, 'text', true, false, false,
    '{"max_length": 100}', 'e.g., Senior Agent', 'Professional designation or title')
ON CONFLICT (field_name) DO NOTHING;

-- Seed select options for dropdown fields
UPDATE agent_profile_field_metadata
SET select_options = '[
    {"value": "MR", "label": "Mr."},
    {"value": "MRS", "label": "Mrs."},
    {"value": "MS", "label": "Ms."},
    {"value": "DR", "label": "Dr."}
]'::jsonb
WHERE field_name = 'title';

UPDATE agent_profile_field_metadata
SET select_options = '[
    {"value": "MALE", "label": "Male"},
    {"value": "FEMALE", "label": "Female"},
    {"value": "OTHER", "label": "Other"}
]'::jsonb
WHERE field_name = 'gender';

UPDATE agent_profile_field_metadata
SET select_options = '[
    {"value": "SINGLE", "label": "Single"},
    {"value": "MARRIED", "label": "Married"},
    {"value": "DIVORCED", "label": "Divorced"},
    {"value": "WIDOWED", "label": "Widowed"}
]'::jsonb
WHERE field_name = 'marital_status';

UPDATE agent_profile_field_metadata
SET select_options = '[
    {"value": "GENERAL", "label": "General"},
    {"value": "OBC", "label": "OBC"},
    {"value": "SC", "label": "SC"},
    {"value": "ST", "label": "ST"}
]'::jsonb
WHERE field_name = 'category';

-- Seed data for contact section
INSERT INTO agent_profile_field_metadata (
    field_name, section, display_name, display_order, field_type,
    is_editable, is_required, requires_approval, validation_rules, placeholder, help_text
) VALUES
('mobile_number', 'contact', 'Mobile Number', 1, 'phone', true, true, false,
    '{"length": 10, "pattern": "^[6-9][0-9]{9}$"}',
    '9876543210', 'Primary mobile number - 10 digits'),
('alternate_number', 'contact', 'Alternate Number', 2, 'phone', true, false, false,
    '{"length": 10, "pattern": "^[6-9][0-9]{9}$"}',
    '9876543210', 'Alternate contact number (optional)'),
('whatsapp_number', 'contact', 'WhatsApp Number', 3, 'phone', true, false, false,
    '{"length": 10, "pattern": "^[6-9][0-9]{9}$"}',
    '9876543210', 'WhatsApp number (optional)')
ON CONFLICT (field_name) DO NOTHING;

-- Seed data for email section
INSERT INTO agent_profile_field_metadata (
    field_name, section, display_name, display_order, field_type,
    is_editable, is_required, requires_approval, validation_rules, placeholder, help_text
) VALUES
('email_address', 'email', 'Email Address', 1, 'email', true, true, false,
    '{"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"}',
    'agent@example.com', 'Primary email address'),
('alternate_email', 'email', 'Alternate Email', 2, 'email', true, false, false,
    '{"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"}',
    'alternate@example.com', 'Alternate email address (optional)')
ON CONFLICT (field_name) DO NOTHING;

-- Seed data for bank section
INSERT INTO agent_profile_field_metadata (
    field_name, section, display_name, display_order, field_type,
    is_editable, is_required, requires_approval, validation_rules, placeholder, help_text
) VALUES
('bank_name', 'bank', 'Bank Name', 1, 'text', true, true, true,
    '{"min_length": 3, "max_length": 100}',
    'State Bank of India', 'Name of the bank'),
('account_number', 'bank', 'Account Number', 2, 'text', true, true, true,
    '{"min_length": 9, "max_length": 18, "pattern": "^[0-9]+$"}',
    '12345678901234', 'Bank account number'),
('ifsc_code', 'bank', 'IFSC Code', 3, 'text', true, true, true,
    '{"length": 11, "pattern": "^[A-Z]{4}0[A-Z0-9]{6}$"}',
    'SBIN0001234', 'Bank IFSC code - 11 characters'),
('account_holder_name', 'bank', 'Account Holder Name', 4, 'text', true, true, true,
    '{"min_length": 2, "max_length": 100}',
    'John Doe', 'Name as per bank account'),
('branch_name', 'bank', 'Branch Name', 5, 'text', true, false, false,
    '{"max_length": 100}',
    'Mumbai Branch', 'Bank branch name (optional)')
ON CONFLICT (field_name) DO NOTHING;

-- Seed data for address section
INSERT INTO agent_profile_field_metadata (
    field_name, section, display_name, display_order, field_type,
    is_editable, is_required, requires_approval, validation_rules, placeholder, help_text
) VALUES
('address_line1', 'address', 'Address Line 1', 1, 'text', true, true, false,
    '{"min_length": 5, "max_length": 200}',
    'House/Flat No., Building Name', 'Primary address line'),
('address_line2', 'address', 'Address Line 2', 2, 'text', true, false, false,
    '{"max_length": 200}',
    'Street, Locality', 'Secondary address line (optional)'),
('city', 'address', 'City', 3, 'text', true, true, false,
    '{"min_length": 2, "max_length": 100}',
    'Mumbai', 'City name'),
('state', 'address', 'State', 4, 'select', true, true, false,
    '{}',
    'Select state', 'State/UT'),
('pincode', 'address', 'Pincode', 5, 'text', true, true, false,
    '{"length": 6, "pattern": "^[0-9]{6}$"}',
    '400001', 'Area pincode - 6 digits'),
('country', 'address', 'Country', 6, 'text', true, true, false,
    '{}',
    'India', 'Country (default: India)')
ON CONFLICT (field_name) DO NOTHING;

-- Update state field with select options (Indian states)
UPDATE agent_profile_field_metadata
SET select_options = '[
    {"value": "AN", "label": "Andaman and Nicobar Islands"},
    {"value": "AP", "label": "Andhra Pradesh"},
    {"value": "AR", "label": "Arunachal Pradesh"},
    {"value": "AS", "label": "Assam"},
    {"value": "BR", "label": "Bihar"},
    {"value": "CH", "label": "Chandigarh"},
    {"value": "CT", "label": "Chhattisgarh"},
    {"value": "DN", "label": "Dadra and Nagar Haveli"},
    {"value": "DD", "label": "Daman and Diu"},
    {"value": "DL", "label": "Delhi"},
    {"value": "GA", "label": "Goa"},
    {"value": "GJ", "label": "Gujarat"},
    {"value": "HR", "label": "Haryana"},
    {"value": "HP", "label": "Himachal Pradesh"},
    {"value": "JK", "label": "Jammu and Kashmir"},
    {"value": "JH", "label": "Jharkhand"},
    {"value": "KA", "label": "Karnataka"},
    {"value": "KL", "label": "Kerala"},
    {"value": "LA", "label": "Ladakh"},
    {"value": "LD", "label": "Lakshadweep"},
    {"value": "MP", "label": "Madhya Pradesh"},
    {"value": "MH", "label": "Maharashtra"},
    {"value": "MN", "label": "Manipur"},
    {"value": "ML", "label": "Meghalaya"},
    {"value": "MZ", "label": "Mizoram"},
    {"value": "NL", "label": "Nagaland"},
    {"value": "OR", "label": "Odisha"},
    {"value": "PY", "label": "Puducherry"},
    {"value": "PB", "label": "Punjab"},
    {"value": "RJ", "label": "Rajasthan"},
    {"value": "SK", "label": "Sikkim"},
    {"value": "TN", "label": "Tamil Nadu"},
    {"value": "TG", "label": "Telangana"},
    {"value": "TR", "label": "Tripura"},
    {"value": "UP", "label": "Uttar Pradesh"},
    {"value": "UT", "label": "Uttarakhand"},
    {"value": "WB", "label": "West Bengal"}
]'::jsonb
WHERE field_name = 'state';

-- Comments
COMMENT ON TABLE agent_profile_field_metadata IS 'Stores metadata about profile fields for dynamic form generation and validation';
COMMENT ON COLUMN agent_profile_field_metadata.validation_rules IS 'JSONB containing validation rules: min_length, max_length, pattern, min, max, etc.';
COMMENT ON COLUMN agent_profile_field_metadata.select_options IS 'JSONB array of {value, label} objects for dropdown fields';
COMMENT ON COLUMN agent_profile_field_metadata.requires_approval IS 'If true, changes to this field require approval before being applied';
