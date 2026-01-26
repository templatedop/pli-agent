-- ============================================================================
-- Agent Profile Management - PostgreSQL Database Schema
-- ============================================================================
-- Version: 1.0
-- Date: 2026-01-25
-- PostgreSQL Version: 16
-- Module: Agent Profile Management
-- Team: Team 1 - Agent Management
-- ============================================================================

-- ============================================================================
-- SECTION 1: EXTENSIONS
-- ============================================================================

-- Create extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create extension for encryption (for bank account numbers)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- SECTION 2: ENUM TYPES
-- ============================================================================

-- Agent Types (BR-AGT-PRF-001, BR-AGT-PRF-002, BR-AGT-PRF-003, BR-AGT-PRF-004)
CREATE TYPE agent_type_enum AS ENUM (
    'ADVISOR',
    'ADVISOR_COORDINATOR',
    'DEPARTMENTAL_EMPLOYEE',
    'FIELD_OFFICER'
);

-- Agent Status (BR-AGT-PRF-016, BR-AGT-PRF-017)
CREATE TYPE agent_status_enum AS ENUM (
    'ACTIVE',
    'SUSPENDED',
    'TERMINATED',
    'DEACTIVATED',
    'EXPIRED'
);

-- Gender (VR-AGT-PRF-005)
CREATE TYPE gender_enum AS ENUM ('Male', 'Female', 'Other');

-- Marital Status (VR-AGT-PRF-006)
CREATE TYPE marital_status_enum AS ENUM (
    'Single',
    'Married',
    'Widowed',
    'Divorced'
);

-- Address Types (BR-AGT-PRF-008)
CREATE TYPE address_type_enum AS ENUM (
    'OFFICIAL',
    'PERMANENT',
    'COMMUNICATION'
);

-- Contact Types (VR-AGT-PRF-011)
CREATE TYPE contact_type_enum AS ENUM (
    'MOBILE',
    'OFFICIAL_LANDLINE',
    'RESIDENT_LANDLINE'
);

-- Email Types (VR-AGT-PRF-012)
CREATE TYPE email_type_enum AS ENUM (
    'OFFICIAL',
    'PERMANENT',
    'COMMUNICATION'
);

-- Account Types (VR-AGT-PRF-017)
CREATE TYPE account_type_enum AS ENUM ('SAVINGS', 'CURRENT');

-- License Line
CREATE TYPE license_line_enum AS ENUM ('LIFE');

-- License Types (BR-AGT-PRF-012)
CREATE TYPE license_type_enum AS ENUM ('PROVISIONAL', 'PERMANENT');

-- Resident Status
CREATE TYPE resident_status_enum AS ENUM ('RESIDENT', 'NON_RESIDENT');

-- License Status (BR-AGT-PRF-012, BR-AGT-PRF-013)
CREATE TYPE license_status_enum AS ENUM ('ACTIVE', 'EXPIRED', 'RENEWED');

-- Reminder Types (BR-AGT-PRF-014)
CREATE TYPE reminder_type_enum AS ENUM (
    '30_DAYS',
    '15_DAYS',
    '7_DAYS',
    'EXPIRY_DAY'
);

-- Reminder Status
CREATE TYPE reminder_status_enum AS ENUM ('PENDING', 'SENT', 'FAILED');

-- Audit Action Types
CREATE TYPE audit_action_enum AS ENUM (
    'CREATE',
    'UPDATE',
    'DELETE',
    'STATUS_CHANGE',
    'LICENSE_ADD',
    'LICENSE_UPDATE',
    'BANK_UPDATE',
    'ADDRESS_UPDATE',
    'CONTACT_UPDATE',
    'EMAIL_UPDATE',
    'LOGIN',
    'LOGOUT',
    'TERMINATE',
    'ACTIVATE',
    'SUSPEND'
);

-- ============================================================================
-- SECTION 3: TABLES
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Table: agent_profiles (E-01: Agent Profile Entity)
-- ----------------------------------------------------------------------------
CREATE TABLE agent_profiles (
    -- Primary Key
    agent_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Profile Identification
    agent_code VARCHAR(20) UNIQUE,
    agent_type agent_type_enum NOT NULL,
    employee_id VARCHAR(20) UNIQUE,

    -- Office and Hierarchy (BR-AGT-PRF-002)
    office_code VARCHAR(20) NOT NULL,
    circle_id VARCHAR(50),
    division_id VARCHAR(50),
    advisor_coordinator_id UUID REFERENCES agent_profiles(agent_id) ON DELETE SET NULL,

    -- Personal Information (VR-AGT-PRF-001 to VR-AGT-PRF-007)
    title VARCHAR(10),
    first_name VARCHAR(50) NOT NULL,
    middle_name VARCHAR(50),
    last_name VARCHAR(50) NOT NULL,
    gender gender_enum NOT NULL,
    date_of_birth DATE NOT NULL,
    category VARCHAR(50),
    marital_status marital_status_enum,

    -- Identification Numbers (VR-AGT-PRF-003, VR-AGT-PRF-004)
    aadhar_number VARCHAR(12) UNIQUE,
    pan_number VARCHAR(10) NOT NULL UNIQUE,

    -- Professional Information
    designation_rank VARCHAR(50),
    service_number VARCHAR(20),
    professional_title VARCHAR(50),

    -- Status Management (BR-AGT-PRF-016, BR-AGT-PRF-017)
    status agent_status_enum NOT NULL DEFAULT 'ACTIVE',
    status_date DATE NOT NULL DEFAULT CURRENT_DATE,
    status_reason VARCHAR(200),

    -- Distribution Channel and Product Authorization (BR-AGT-PRF-026)
    distribution_channel VARCHAR(50),
    product_class VARCHAR(50),
    external_identification_number VARCHAR(50),

    -- Goals and Performance (BR-AGT-PRF-024)
    goals JSONB,

    -- Workflow State Management
    workflow_state VARCHAR(50),
    workflow_state_history JSONB,

    -- Metadata and Search
    metadata JSONB,
    search_vector tsvector,

    -- Audit Fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    created_by VARCHAR(50) NOT NULL,
    updated_by VARCHAR(50),
    deleted_at TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 0
);

-- ----------------------------------------------------------------------------
-- Table: agent_addresses (E-02: Agent Address Entity)
-- ----------------------------------------------------------------------------
CREATE TABLE agent_addresses (
    -- Primary Key
    address_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign Key
    agent_id UUID NOT NULL REFERENCES agent_profiles(agent_id) ON DELETE CASCADE,

    -- Address Type (BR-AGT-PRF-008)
    address_type address_type_enum NOT NULL,

    -- Address Fields (VR-AGT-PRF-007, VR-AGT-PRF-008)
    address_line1 VARCHAR(200) NOT NULL,
    address_line2 VARCHAR(200),
    village VARCHAR(50),
    taluka VARCHAR(50),
    city VARCHAR(50) NOT NULL,
    district VARCHAR(50),
    state VARCHAR(50) NOT NULL,
    country VARCHAR(50) NOT NULL DEFAULT 'India',
    pincode VARCHAR(6) NOT NULL,

    -- Communication Address Flag (BR-AGT-PRF-009)
    is_same_as_permanent BOOLEAN DEFAULT FALSE,

    -- Effective Date
    effective_from DATE NOT NULL DEFAULT CURRENT_DATE,

    -- Metadata
    metadata JSONB,

    -- Audit Fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    created_by VARCHAR(50) NOT NULL,
    updated_by VARCHAR(50),
    deleted_at TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 0
);

-- ----------------------------------------------------------------------------
-- Table: agent_contacts (E-03: Agent Contact Entity)
-- ----------------------------------------------------------------------------
CREATE TABLE agent_contacts (
    -- Primary Key
    contact_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign Key
    agent_id UUID NOT NULL REFERENCES agent_profiles(agent_id) ON DELETE CASCADE,

    -- Contact Type (BR-AGT-PRF-010)
    contact_type contact_type_enum NOT NULL,

    -- Contact Number (VR-AGT-PRF-011)
    contact_number VARCHAR(15) NOT NULL,

    -- Primary Contact Flag
    is_primary BOOLEAN DEFAULT FALSE,

    -- Effective Date
    effective_from DATE NOT NULL DEFAULT CURRENT_DATE,

    -- Metadata
    metadata JSONB,

    -- Audit Fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    created_by VARCHAR(50) NOT NULL,
    updated_by VARCHAR(50),
    deleted_at TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 0
);

-- ----------------------------------------------------------------------------
-- Table: agent_emails (E-04: Agent Email Entity)
-- ----------------------------------------------------------------------------
CREATE TABLE agent_emails (
    -- Primary Key
    email_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign Key
    agent_id UUID NOT NULL REFERENCES agent_profiles(agent_id) ON DELETE CASCADE,

    -- Email Type (BR-AGT-PRF-011)
    email_type email_type_enum NOT NULL,

    -- Email Address (VR-AGT-PRF-012)
    email_address VARCHAR(100) NOT NULL,

    -- Primary Email Flag
    is_primary BOOLEAN DEFAULT FALSE,

    -- Effective Date
    effective_from DATE NOT NULL DEFAULT CURRENT_DATE,

    -- Metadata
    metadata JSONB,

    -- Audit Fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    created_by VARCHAR(50) NOT NULL,
    updated_by VARCHAR(50),
    deleted_at TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 0
);

-- ----------------------------------------------------------------------------
-- Table: agent_bank_details (E-05: Agent Bank Details Entity)
-- ----------------------------------------------------------------------------
CREATE TABLE agent_bank_details (
    -- Primary Key
    bank_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign Key
    agent_id UUID NOT NULL REFERENCES agent_profiles(agent_id) ON DELETE CASCADE,

    -- Account Type (VR-AGT-PRF-017)
    account_type account_type_enum NOT NULL,

    -- Bank Details (BR-AGT-PRF-018, VR-AGT-PRF-015, VR-AGT-PRF-016)
    -- Encrypted account number using pgcrypto
    account_number_encrypted BYTEA NOT NULL,
    ifsc_code VARCHAR(11) NOT NULL,
    bank_name VARCHAR(100) NOT NULL,
    branch_name VARCHAR(100),

    -- Effective Date
    effective_from DATE NOT NULL DEFAULT CURRENT_DATE,

    -- Metadata
    metadata JSONB,

    -- Audit Fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    created_by VARCHAR(50) NOT NULL,
    updated_by VARCHAR(50),
    deleted_at TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 0
);

-- ----------------------------------------------------------------------------
-- Table: agent_licenses (E-06: Agent License Entity)
-- ----------------------------------------------------------------------------
CREATE TABLE agent_licenses (
    -- Primary Key
    license_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign Key
    agent_id UUID NOT NULL REFERENCES agent_profiles(agent_id) ON DELETE CASCADE,

    -- License Details (BR-AGT-PRF-012, VR-AGT-PRF-013, VR-AGT-PRF-014)
    license_line license_line_enum NOT NULL DEFAULT 'LIFE',
    license_type license_type_enum NOT NULL,
    license_number VARCHAR(30) NOT NULL UNIQUE,
    resident_status resident_status_enum NOT NULL DEFAULT 'RESIDENT',

    -- License Dates (BR-AGT-PRF-030)
    license_date DATE NOT NULL,
    renewal_date DATE NOT NULL,
    authority_date DATE NOT NULL,

    -- Renewal Tracking (BR-AGT-PRF-012)
    renewal_count INTEGER NOT NULL DEFAULT 0,
    license_status license_status_enum NOT NULL DEFAULT 'ACTIVE',

    -- License Exam Status (BR-AGT-PRF-012)
    licentiate_exam_passed BOOLEAN DEFAULT FALSE,
    licentiate_exam_date DATE,
    licentiate_certificate_number VARCHAR(30),

    -- Primary License Flag
    is_primary BOOLEAN DEFAULT FALSE,

    -- Metadata
    metadata JSONB,

    -- Audit Fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    created_by VARCHAR(50) NOT NULL,
    updated_by VARCHAR(50),
    deleted_at TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 0
);

-- ----------------------------------------------------------------------------
-- Table: agent_license_reminders (E-07: Agent License Reminder Log Entity)
-- ----------------------------------------------------------------------------
CREATE TABLE agent_license_reminders (
    -- Primary Key
    reminder_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign Key
    license_id UUID NOT NULL REFERENCES agent_licenses(license_id) ON DELETE CASCADE,

    -- Reminder Details (BR-AGT-PRF-014)
    reminder_type reminder_type_enum NOT NULL,
    reminder_date DATE NOT NULL,
    sent_date TIMESTAMP,

    -- Sent Status
    sent_status reminder_status_enum NOT NULL DEFAULT 'PENDING',

    -- Sent Flags
    email_sent BOOLEAN DEFAULT FALSE,
    sms_sent BOOLEAN DEFAULT FALSE,

    -- Failure Tracking
    failure_reason VARCHAR(200),
    retry_count INTEGER DEFAULT 0,

    -- Metadata
    metadata JSONB,

    -- Audit Fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(50) NOT NULL
);

-- ----------------------------------------------------------------------------
-- Table: agent_audit_logs (E-08: Agent Audit Log Entity)
-- ----------------------------------------------------------------------------
CREATE TABLE agent_audit_logs (
    -- Primary Key
    audit_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Foreign Key
    agent_id UUID NOT NULL REFERENCES agent_profiles(agent_id) ON DELETE CASCADE,

    -- Action Details (BR-AGT-PRF-005, BR-AGT-PRF-006)
    action_type audit_action_enum NOT NULL,
    field_name VARCHAR(50),
    old_value TEXT,
    new_value TEXT,
    action_reason VARCHAR(500),

    -- Performed By (BR-AGT-PRF-005)
    performed_by VARCHAR(50) NOT NULL,
    performed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- IP Address for Security
    ip_address VARCHAR(45),

    -- Metadata
    metadata JSONB,

    -- Audit Fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- SECTION 4: CHECK CONSTRAINTS
-- ============================================================================

-- PAN Format Validation (VR-AGT-PRF-003: BR-AGT-PRF-006)
ALTER TABLE agent_profiles
ADD CONSTRAINT chk_agent_profiles_pan_format
CHECK (pan_number ~ '^[A-Z]{5}[0-9]{4}[A-Z]{1}$');

-- Age Validation (VR-AGT-PRF-006)
ALTER TABLE agent_profiles
ADD CONSTRAINT chk_agent_profiles_age_range
CHECK (date_of_birth BETWEEN (CURRENT_DATE - INTERVAL '70 years') AND (CURRENT_DATE - INTERVAL '18 years'));

-- Aadhar Format Validation (VR-AGT-PRF-004)
ALTER TABLE agent_profiles
ADD CONSTRAINT chk_agent_profiles_aadhar_format
CHECK (aadhar_number IS NULL OR aadhar_number ~ '^[0-9]{12}$');

-- Advisor Coordinator Linkage (BR-AGT-PRF-001)
ALTER TABLE agent_profiles
ADD CONSTRAINT chk_agent_profiles_advisor_coordinator_required
CHECK (
    (agent_type != 'ADVISOR') OR
    (advisor_coordinator_id IS NOT NULL)
);

-- Status Reason Mandatory (BR-AGT-PRF-016)
ALTER TABLE agent_profiles
ADD CONSTRAINT chk_agent_profiles_status_reason_required
CHECK (
    (status NOT IN ('SUSPENDED', 'TERMINATED', 'DEACTIVATED')) OR
    (status_reason IS NOT NULL AND LENGTH(status_reason) >= 10)
);

-- Coordinator Geographic Assignment (BR-AGT-PRF-002)
ALTER TABLE agent_profiles
ADD CONSTRAINT chk_agent_profiles_coordinator_geographic
CHECK (
    (agent_type != 'ADVISOR_COORDINATOR') OR
    (circle_id IS NOT NULL AND division_id IS NOT NULL)
);

-- Pincode Format Validation (VR-AGT-PRF-008)
ALTER TABLE agent_addresses
ADD CONSTRAINT chk_agent_addresses_pincode_format
CHECK (pincode ~ '^[0-9]{6}$');

-- Contact Number Format Validation (VR-AGT-PRF-011)
ALTER TABLE agent_contacts
ADD CONSTRAINT chk_agent_contacts_mobile_format
CHECK (
    (contact_type = 'MOBILE' AND contact_number ~ '^[0-9]{10}$') OR
    (contact_type IN ('OFFICIAL_LANDLINE', 'RESIDENT_LANDLINE'))
);

-- Email Format Validation (VR-AGT-PRF-012)
ALTER TABLE agent_emails
ADD CONSTRAINT chk_agent_emails_format
CHECK (email_address ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');

-- IFSC Format Validation (VR-AGT-PRF-016)
ALTER TABLE agent_bank_details
ADD CONSTRAINT chk_agent_bank_details_ifsc_format
CHECK (ifsc_code ~ '^[A-Z]{4}[0-9]{7}$');

-- License Renewal Date Validation (BR-AGT-PRF-030)
ALTER TABLE agent_licenses
ADD CONSTRAINT chk_agent_licenses_renewal_after_license
CHECK (renewal_date >= license_date);

-- License Renewal Count Validation (BR-AGT-PRF-012)
ALTER TABLE agent_licenses
ADD CONSTRAINT chk_agent_licenses_provisional_renewal_limit
CHECK (
    (license_type != 'PROVISIONAL') OR
    (renewal_count <= 2)
);

-- ============================================================================
-- SECTION 5: UNIQUE CONSTRAINTS (Additional to column-level UNIQUE)
-- ============================================================================

-- Ensure one primary contact per agent per contact type
CREATE UNIQUE INDEX idx_agent_contacts_primary_contact
ON agent_contacts (agent_id, contact_type)
WHERE is_primary = TRUE;

-- Ensure one primary email per agent per email type
CREATE UNIQUE INDEX idx_agent_emails_primary_email
ON agent_emails (agent_id, email_type)
WHERE is_primary = TRUE;

-- Ensure unique bank account per agent
CREATE UNIQUE INDEX idx_agent_bank_details_unique_account
ON agent_bank_details (agent_id, account_number_encrypted, ifsc_code);

-- ============================================================================
-- SECTION 6: INDEXES
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Foreign Key Indexes
-- ----------------------------------------------------------------------------
CREATE INDEX idx_agent_addresses_agent_id ON agent_addresses(agent_id);
CREATE INDEX idx_agent_contacts_agent_id ON agent_contacts(agent_id);
CREATE INDEX idx_agent_emails_agent_id ON agent_emails(agent_id);
CREATE INDEX idx_agent_bank_details_agent_id ON agent_bank_details(agent_id);
CREATE INDEX idx_agent_licenses_agent_id ON agent_licenses(agent_id);
CREATE INDEX idx_agent_license_reminders_license_id ON agent_license_reminders(license_id);
CREATE INDEX idx_agent_audit_logs_agent_id ON agent_audit_logs(agent_id);
CREATE INDEX idx_agent_profiles_advisor_coordinator_id ON agent_profiles(advisor_coordinator_id);

-- ----------------------------------------------------------------------------
-- Status and Workflow State Indexes
-- ----------------------------------------------------------------------------
CREATE INDEX idx_agent_profiles_status ON agent_profiles(status);
CREATE INDEX idx_agent_profiles_workflow_state ON agent_profiles(workflow_state);
CREATE INDEX idx_agent_licenses_license_status ON agent_licenses(license_status);

-- ----------------------------------------------------------------------------
-- Date Indexes
-- ----------------------------------------------------------------------------
CREATE INDEX idx_agent_profiles_created_at ON agent_profiles(created_at);
CREATE INDEX idx_agent_profiles_date_of_birth ON agent_profiles(date_of_birth);
CREATE INDEX idx_agent_licenses_license_date ON agent_licenses(license_date);
CREATE INDEX idx_agent_licenses_renewal_date ON agent_licenses(renewal_date);
CREATE INDEX idx_agent_license_reminders_reminder_date ON agent_license_reminders(reminder_date);
CREATE INDEX idx_agent_audit_logs_performed_at ON agent_audit_logs(performed_at);

-- ----------------------------------------------------------------------------
-- Composite Indexes
-- ----------------------------------------------------------------------------
CREATE INDEX idx_agent_profiles_type_status ON agent_profiles(agent_type, status);
CREATE INDEX idx_agent_profiles_office_status ON agent_profiles(office_code, status);
CREATE INDEX idx_agent_contacts_agent_primary ON agent_contacts(agent_id, is_primary);
CREATE INDEX idx_agent_emails_agent_primary ON agent_emails(agent_id, is_primary);
CREATE INDEX idx_agent_licenses_agent_status ON agent_licenses(agent_id, license_status);
CREATE INDEX idx_agent_audit_logs_agent_performed_at ON agent_audit_logs(agent_id, performed_at);

-- ----------------------------------------------------------------------------
-- GIN Indexes for JSONB and Full-Text Search
-- ----------------------------------------------------------------------------
CREATE INDEX idx_agent_profiles_metadata ON agent_profiles USING GIN (metadata);
CREATE INDEX idx_agent_profiles_search_vector ON agent_profiles USING GIN (search_vector);
CREATE INDEX idx_agent_addresses_metadata ON agent_addresses USING GIN (metadata);
CREATE INDEX idx_agent_contacts_metadata ON agent_contacts USING GIN (metadata);
CREATE INDEX idx_agent_emails_metadata ON agent_emails USING GIN (metadata);
CREATE INDEX idx_agent_bank_details_metadata ON agent_bank_details USING GIN (metadata);
CREATE INDEX idx_agent_licenses_metadata ON agent_licenses USING GIN (metadata);
CREATE INDEX idx_agent_audit_logs_metadata ON agent_audit_logs USING GIN (metadata);

-- ----------------------------------------------------------------------------
-- Partial Indexes for Performance
-- ----------------------------------------------------------------------------
CREATE INDEX idx_agent_profiles_active ON agent_profiles(status)
WHERE status = 'ACTIVE' AND deleted_at IS NULL;

CREATE INDEX idx_agent_licenses_active_renewals ON agent_licenses(renewal_date)
WHERE license_status = 'ACTIVE';

CREATE INDEX idx_agent_license_reminders_pending ON agent_license_reminders(reminder_date)
WHERE sent_status = 'PENDING';

-- ============================================================================
-- SECTION 7: FUNCTIONS
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Function: fn_update_updated_at_column
-- Purpose: Auto-update updated_at timestamp on row update
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- Function: fn_update_search_vector
-- Purpose: Update full-text search vector for agent profiles
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_update_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('simple', COALESCE(NEW.first_name, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(NEW.middle_name, '')), 'B') ||
        setweight(to_tsvector('simple', COALESCE(NEW.last_name, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(NEW.pan_number, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(NEW.agent_code, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(NEW.employee_id, '')), 'B') ||
        setweight(to_tsvector('simple', COALESCE(NEW.designation_rank, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- Function: fn_send_license_renewal_reminders
-- Purpose: Scheduled job to send license renewal reminders (BR-AGT-PRF-014)
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_send_license_renewal_reminders()
RETURNS INTEGER AS $$
DECLARE
    v_count INTEGER := 0;
    reminder_rec RECORD;
BEGIN
    -- Find licenses needing reminders
    FOR reminder_rec IN
        SELECT
            al.license_id,
            al.agent_id,
            al.renewal_date,
            CASE
                WHEN CURRENT_DATE = (al.renewal_date - INTERVAL '30 days') THEN '30_DAYS'::reminder_type_enum
                WHEN CURRENT_DATE = (al.renewal_date - INTERVAL '15 days') THEN '15_DAYS'::reminder_type_enum
                WHEN CURRENT_DATE = (al.renewal_date - INTERVAL '7 days') THEN '7_DAYS'::reminder_type_enum
                WHEN CURRENT_DATE = al.renewal_date THEN 'EXPIRY_DAY'::reminder_type_enum
            END AS reminder_type
        FROM agent_licenses al
        WHERE al.license_status = 'ACTIVE'
          AND al.deleted_at IS NULL
          AND (
              CURRENT_DATE = (al.renewal_date - INTERVAL '30 days') OR
              CURRENT_DATE = (al.renewal_date - INTERVAL '15 days') OR
              CURRENT_DATE = (al.renewal_date - INTERVAL '7 days') OR
              CURRENT_DATE = al.renewal_date
          )
    LOOP
        -- Insert reminder record
        INSERT INTO agent_license_reminders (
            license_id,
            reminder_type,
            reminder_date,
            sent_status,
            created_by
        ) VALUES (
            reminder_rec.license_id,
            reminder_rec.reminder_type,
            CURRENT_DATE,
            'PENDING',
            'SYSTEM'
        );

        v_count := v_count + 1;
    END LOOP;

    RETURN v_count;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- Function: fn_auto_deactivate_expired_licenses
-- Purpose: Scheduled job to deactivate agents with expired licenses (BR-AGT-PRF-013)
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_auto_deactivate_expired_licenses()
RETURNS INTEGER AS $$
DECLARE
    v_count INTEGER := 0;
    agent_rec RECORD;
BEGIN
    -- Find agents with expired licenses
    FOR agent_rec IN
        SELECT DISTINCT
            ap.agent_id,
            ap.agent_code,
            ap.first_name,
            ap.last_name,
            al.license_id,
            al.renewal_date
        FROM agent_profiles ap
        JOIN agent_licenses al ON ap.agent_id = al.agent_id
        WHERE ap.status = 'ACTIVE'
          AND ap.deleted_at IS NULL
          AND al.license_status = 'ACTIVE'
          AND al.renewal_date < CURRENT_DATE
    LOOP
        -- Update agent status
        UPDATE agent_profiles
        SET
            status = 'DEACTIVATED',
            status_date = CURRENT_DATE,
            status_reason = 'License expired on ' || agent_rec.renewal_date,
            updated_at = CURRENT_TIMESTAMP
        WHERE agent_id = agent_rec.agent_id;

        -- Update license status
        UPDATE agent_licenses
        SET
            license_status = 'EXPIRED',
            updated_at = CURRENT_TIMESTAMP
        WHERE license_id = agent_rec.license_id;

        -- Log audit trail
        INSERT INTO agent_audit_logs (
            agent_id,
            action_type,
            field_name,
            old_value,
            new_value,
            action_reason,
            performed_by,
            performed_at
        ) VALUES (
            agent_rec.agent_id,
            'STATUS_CHANGE',
            'status',
            'ACTIVE',
            'DEACTIVATED',
            'License expired on ' || agent_rec.renewal_date,
            'SYSTEM',
            CURRENT_TIMESTAMP
        );

        v_count := v_count + 1;
    END LOOP;

    RETURN v_count;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- Function: fn_validate_workflow_transition
-- Purpose: Validate workflow state transitions
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_validate_workflow_transition(
    p_table_name TEXT,
    p_record_id UUID,
    p_old_state TEXT,
    p_new_state TEXT
)
RETURNS BOOLEAN AS $$
DECLARE
    v_valid_transition BOOLEAN := FALSE;
BEGIN
    -- Define valid state transitions
    -- This is a placeholder for actual workflow logic
    -- Implement based on Temporal workflow definitions

    IF p_old_state IS NULL THEN
        -- Initial state
        v_valid_transition := TRUE;
    ELSE
        -- Check if transition is valid
        -- Add specific state machine logic here
        v_valid_transition := TRUE;
    END IF;

    RETURN v_valid_transition;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- SECTION 8: TRIGGERS
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Trigger: agent_profiles_update_updated_at
-- Purpose: Auto-update updated_at and version for agent_profiles
-- ----------------------------------------------------------------------------
CREATE TRIGGER agent_profiles_update_updated_at
BEFORE UPDATE ON agent_profiles
FOR EACH ROW
EXECUTE FUNCTION fn_update_updated_at_column();

-- ----------------------------------------------------------------------------
-- Trigger: agent_profiles_update_search_vector
-- Purpose: Update full-text search vector on insert/update
-- ----------------------------------------------------------------------------
CREATE TRIGGER agent_profiles_update_search_vector
BEFORE INSERT OR UPDATE ON agent_profiles
FOR EACH ROW
EXECUTE FUNCTION fn_update_search_vector();

-- ----------------------------------------------------------------------------
-- Trigger: agent_addresses_update_updated_at
-- Purpose: Auto-update updated_at and version for agent_addresses
-- ----------------------------------------------------------------------------
CREATE TRIGGER agent_addresses_update_updated_at
BEFORE UPDATE ON agent_addresses
FOR EACH ROW
EXECUTE FUNCTION fn_update_updated_at_column();

-- ----------------------------------------------------------------------------
-- Trigger: agent_contacts_update_updated_at
-- Purpose: Auto-update updated_at and version for agent_contacts
-- ----------------------------------------------------------------------------
CREATE TRIGGER agent_contacts_update_updated_at
BEFORE UPDATE ON agent_contacts
FOR EACH ROW
EXECUTE FUNCTION fn_update_updated_at_column();

-- ----------------------------------------------------------------------------
-- Trigger: agent_emails_update_updated_at
-- Purpose: Auto-update updated_at and version for agent_emails
-- ----------------------------------------------------------------------------
CREATE TRIGGER agent_emails_update_updated_at
BEFORE UPDATE ON agent_emails
FOR EACH ROW
EXECUTE FUNCTION fn_update_updated_at_column();

-- ----------------------------------------------------------------------------
-- Trigger: agent_bank_details_update_updated_at
-- Purpose: Auto-update updated_at and version for agent_bank_details
-- ----------------------------------------------------------------------------
CREATE TRIGGER agent_bank_details_update_updated_at
BEFORE UPDATE ON agent_bank_details
FOR EACH ROW
EXECUTE FUNCTION fn_update_updated_at_column();

-- ----------------------------------------------------------------------------
-- Trigger: agent_licenses_update_updated_at
-- Purpose: Auto-update updated_at and version for agent_licenses
-- ----------------------------------------------------------------------------
CREATE TRIGGER agent_licenses_update_updated_at
BEFORE UPDATE ON agent_licenses
FOR EACH ROW
EXECUTE FUNCTION fn_update_updated_at_column();

-- ============================================================================
-- SECTION 9: VIEWS
-- ============================================================================

-- ----------------------------------------------------------------------------
-- View: v_active_agents
-- Purpose: Active agents with key details for common queries
-- ----------------------------------------------------------------------------
CREATE VIEW v_active_agents AS
SELECT
    ap.agent_id,
    ap.agent_code,
    ap.first_name,
    ap.middle_name,
    ap.last_name,
    ap.pan_number,
    ap.agent_type,
    ap.status,
    ap.office_code,
    ap.circle_id,
    ap.division_id,
    al.license_number,
    al.license_type,
    al.renewal_date,
    ae.email_address AS primary_email,
    ac.contact_number AS primary_mobile,
    ap.created_at
FROM agent_profiles ap
LEFT JOIN agent_licenses al ON ap.agent_id = al.agent_id AND al.is_primary = true AND al.deleted_at IS NULL
LEFT JOIN LATERAL (
    SELECT email_address
    FROM agent_emails
    WHERE agent_id = ap.agent_id
      AND is_primary = true
      AND deleted_at IS NULL
    LIMIT 1
) ae ON true
LEFT JOIN LATERAL (
    SELECT contact_number
    FROM agent_contacts
    WHERE agent_id = ap.agent_id
      AND is_primary = true
      AND deleted_at IS NULL
    LIMIT 1
) ac ON true
WHERE ap.status = 'ACTIVE'
  AND ap.deleted_at IS NULL;

-- ----------------------------------------------------------------------------
-- View: v_license_renewal_dashboard
-- Purpose: License renewal tracking with urgency indicators (BR-AGT-PRF-014)
-- ----------------------------------------------------------------------------
CREATE VIEW v_license_renewal_dashboard AS
SELECT
    ap.agent_id,
    ap.agent_code,
    ap.first_name || ' ' || COALESCE(ap.middle_name || ' ', '') || ap.last_name AS agent_name,
    ac.contact_number AS mobile_number,
    ae.email_address AS email_address,
    al.license_id,
    al.license_number,
    al.license_type,
    al.license_date,
    al.renewal_date,
    EXTRACT(DAY FROM (al.renewal_date - CURRENT_DATE))::INTEGER AS days_until_renewal,
    CASE
        WHEN EXTRACT(DAY FROM (al.renewal_date - CURRENT_DATE))::INTEGER <= 0 THEN 'EXPIRED'
        WHEN EXTRACT(DAY FROM (al.renewal_date - CURRENT_DATE))::INTEGER <= 7 THEN 'CRITICAL'
        WHEN EXTRACT(DAY FROM (al.renewal_date - CURRENT_DATE))::INTEGER <= 30 THEN 'WARNING'
        ELSE 'NORMAL'
    END AS renewal_urgency,
    CASE
        WHEN al.license_type = 'PROVISIONAL' AND al.renewal_count >= 2 THEN 'FINAL_RENEWAL'
        WHEN al.license_type = 'PROVISIONAL' THEN 'PROVISIONAL_RENEWAL_' || (al.renewal_count + 1)::TEXT
        ELSE 'ANNUAL_RENEWAL'
    END AS renewal_category
FROM agent_profiles ap
JOIN agent_licenses al ON ap.agent_id = al.agent_id AND al.is_primary = true AND al.deleted_at IS NULL
LEFT JOIN LATERAL (
    SELECT contact_number
    FROM agent_contacts
    WHERE agent_id = ap.agent_id AND contact_type = 'MOBILE' AND is_primary = true AND deleted_at IS NULL
    LIMIT 1
) ac ON true
LEFT JOIN LATERAL (
    SELECT email_address
    FROM agent_emails
    WHERE agent_id = ap.agent_id AND is_primary = true AND deleted_at IS NULL
    LIMIT 1
) ae ON true
WHERE al.license_status = 'ACTIVE'
  AND ap.deleted_at IS NULL
ORDER BY al.renewal_date ASC;

-- ----------------------------------------------------------------------------
-- View: v_agent_hierarchy
-- Purpose: Advisor-Coordinator relationship view (BR-AGT-PRF-001)
-- ----------------------------------------------------------------------------
CREATE VIEW v_agent_hierarchy AS
SELECT
    advisor.agent_id AS advisor_id,
    advisor.agent_code AS advisor_code,
    advisor.first_name || ' ' || COALESCE(advisor.middle_name || ' ', '') || advisor.last_name AS advisor_name,
    advisor.status AS advisor_status,
    coordinator.agent_id AS coordinator_id,
    coordinator.agent_code AS coordinator_code,
    coordinator.first_name || ' ' || COALESCE(coordinator.middle_name || ' ', '') || coordinator.last_name AS coordinator_name,
    coordinator.status AS coordinator_status,
    coordinator.circle_id,
    coordinator.division_id,
    coordinator.office_code AS coordinator_office_code
FROM agent_profiles advisor
INNER JOIN agent_profiles coordinator ON advisor.advisor_coordinator_id = coordinator.agent_id
WHERE advisor.agent_type = 'ADVISOR'
  AND advisor.deleted_at IS NULL
  AND coordinator.deleted_at IS NULL;

-- ----------------------------------------------------------------------------
-- View: v_agent_audit_summary
-- Purpose: Agent audit trail summary
-- ----------------------------------------------------------------------------
CREATE VIEW v_agent_audit_summary AS
SELECT
    ap.agent_id,
    ap.agent_code,
    ap.first_name || ' ' || COALESCE(ap.middle_name || ' ', '') || ap.last_name AS agent_name,
    COUNT(aal.audit_id) AS total_changes,
    MAX(aal.performed_at) AS last_change_at,
    COUNT(CASE WHEN aal.action_type = 'STATUS_CHANGE' THEN 1 END) AS status_changes,
    COUNT(CASE WHEN aal.action_type = 'LICENSE_UPDATE' THEN 1 END) AS license_updates,
    COUNT(CASE WHEN aal.action_type = 'BANK_UPDATE' THEN 1 END) AS bank_updates
FROM agent_profiles ap
LEFT JOIN agent_audit_logs aal ON ap.agent_id = aal.agent_id
WHERE ap.deleted_at IS NULL
GROUP BY
    ap.agent_id,
    ap.agent_code,
    ap.first_name,
    ap.middle_name,
    ap.last_name;

-- ============================================================================
-- SECTION 10: COMMENTS AND DOCUMENTATION
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Table: agent_profiles Comments
-- ----------------------------------------------------------------------------
COMMENT ON TABLE agent_profiles IS 'E-01: Agent Profile Entity - Main agent profile table containing all agent personal and professional information (BR-AGT-PRF-001 to BR-AGT-PRF-030)';

COMMENT ON COLUMN agent_profiles.agent_id IS 'Unique agent identifier (UUID primary key)';
COMMENT ON COLUMN agent_profiles.agent_code IS 'Unique agent code generated by system (format: AGENT-YYYY-NNNNNN)';
COMMENT ON COLUMN agent_profiles.agent_type IS 'Agent type: ADVISOR, ADVISOR_COORDINATOR, DEPARTMENTAL_EMPLOYEE, or FIELD_OFFICER (BR-AGT-PRF-001 to BR-AGT-PRF-004)';
COMMENT ON COLUMN agent_profiles.employee_id IS 'HRMS Employee ID for departmental employees (BR-AGT-PRF-003)';
COMMENT ON COLUMN agent_profiles.office_code IS 'Linked office code (BR-AGT-PRF-002)';
COMMENT ON COLUMN agent_profiles.circle_id IS 'Circle assignment for advisor coordinators (BR-AGT-PRF-002)';
COMMENT ON COLUMN agent_profiles.division_id IS 'Division assignment for advisor coordinators (BR-AGT-PRF-002)';
COMMENT ON COLUMN agent_profiles.advisor_coordinator_id IS 'Foreign key to advisor coordinator (mandatory for ADVISOR type - BR-AGT-PRF-001)';
COMMENT ON COLUMN agent_profiles.pan_number IS 'PAN number with format validation AAAAA9999A (VR-AGT-PRF-003, BR-AGT-PRF-006)';
COMMENT ON COLUMN agent_profiles.aadhar_number IS 'Aadhar number with 12-digit validation (VR-AGT-PRF-004)';
COMMENT ON COLUMN agent_profiles.date_of_birth IS 'Date of birth with age validation (18-70 years - VR-AGT-PRF-006)';
COMMENT ON COLUMN agent_profiles.status IS 'Agent status: ACTIVE, SUSPENDED, TERMINATED, DEACTIVATED, or EXPIRED (BR-AGT-PRF-016, BR-AGT-PRF-017)';
COMMENT ON COLUMN agent_profiles.status_reason IS 'Reason for status change (mandatory for SUSPENDED/TERMINATED/DEACTIVATED - BR-AGT-PRF-016)';
COMMENT ON COLUMN agent_profiles.workflow_state IS 'Current workflow state from Temporal workflows';
COMMENT ON COLUMN agent_profiles.workflow_state_history IS 'Workflow state transition history stored as JSONB';
COMMENT ON COLUMN agent_profiles.metadata IS 'Additional metadata stored as JSONB for flexibility';
COMMENT ON COLUMN agent_profiles.search_vector IS 'Full-text search vector for agent name, PAN, and code searches';
COMMENT ON COLUMN agent_profiles.deleted_at IS 'Soft delete timestamp (NULL for active records)';

-- ----------------------------------------------------------------------------
-- Table: agent_addresses Comments
-- ----------------------------------------------------------------------------
COMMENT ON TABLE agent_addresses IS 'E-02: Agent Address Entity - Multiple address types supported (BR-AGT-PRF-008, BR-AGT-PRF-009)';

COMMENT ON COLUMN agent_addresses.address_type IS 'Address type: OFFICIAL, PERMANENT, or COMMUNICATION (BR-AGT-PRF-008)';
COMMENT ON COLUMN agent_addresses.pincode IS '6-digit PIN code with format validation (VR-AGT-PRF-008)';
COMMENT ON COLUMN agent_addresses.is_same_as_permanent IS 'Flag indicating communication address same as permanent (BR-AGT-PRF-009)';

-- ----------------------------------------------------------------------------
-- Table: agent_contacts Comments
-- ----------------------------------------------------------------------------
COMMENT ON TABLE agent_contacts IS 'E-03: Agent Contact Entity - Phone number management (BR-AGT-PRF-010)';

COMMENT ON COLUMN agent_contacts.contact_type IS 'Contact type: MOBILE, OFFICIAL_LANDLINE, or RESIDENT_LANDLINE (BR-AGT-PRF-010)';
COMMENT ON COLUMN agent_contacts.contact_number IS 'Contact number with 10-digit validation for mobile (VR-AGT-PRF-011)';
COMMENT ON COLUMN agent_contacts.is_primary IS 'Primary contact flag (one primary per contact type per agent)';

-- ----------------------------------------------------------------------------
-- Table: agent_emails Comments
-- ----------------------------------------------------------------------------
COMMENT ON TABLE agent_emails IS 'E-04: Agent Email Entity - Email address management (BR-AGT-PRF-011)';

COMMENT ON COLUMN agent_emails.email_type IS 'Email type: OFFICIAL, PERMANENT, or COMMUNICATION (BR-AGT-PRF-011)';
COMMENT ON COLUMN agent_emails.email_address IS 'Email address with format validation (VR-AGT-PRF-012)';
COMMENT ON COLUMN agent_emails.is_primary IS 'Primary email flag (one primary per email type per agent)';

-- ----------------------------------------------------------------------------
-- Table: agent_bank_details Comments
-- ----------------------------------------------------------------------------
COMMENT ON TABLE agent_bank_details IS 'E-05: Agent Bank Details Entity - Bank/POSB account details for commission disbursement (BR-AGT-PRF-018)';

COMMENT ON COLUMN agent_bank_details.account_number_encrypted IS 'Encrypted bank account number using pgcrypto (VR-AGT-PRF-015)';
COMMENT ON COLUMN agent_bank_details.ifsc_code IS 'IFSC code with format validation AAAA0123456 (VR-AGT-PRF-016)';
COMMENT ON COLUMN agent_bank_details.bank_name IS 'Bank name auto-fetched from IFSC (VR-AGT-PRF-016)';
COMMENT ON COLUMN agent_bank_details.account_type IS 'Account type: SAVINGS or CURRENT (VR-AGT-PRF-017)';

-- ----------------------------------------------------------------------------
-- Table: agent_licenses Comments
-- ----------------------------------------------------------------------------
COMMENT ON TABLE agent_licenses IS 'E-06: Agent License Entity - License details and renewal tracking (BR-AGT-PRF-012, BR-AGT-PRF-013, BR-AGT-PRF-014, BR-AGT-PRF-030)';

COMMENT ON COLUMN agent_licenses.license_type IS 'License type: PROVISIONAL or PERMANENT (BR-AGT-PRF-012)';
COMMENT ON COLUMN agent_licenses.license_number IS 'Unique license number';
COMMENT ON COLUMN agent_licenses.license_date IS 'License issue date (BR-AGT-PRF-030)';
COMMENT ON COLUMN agent_licenses.renewal_date IS 'License renewal date (BR-AGT-PRF-012, BR-AGT-PRF-014)';
COMMENT ON COLUMN agent_licenses.renewal_count IS 'Number of renewals completed (max 2 for provisional - BR-AGT-PRF-012)';
COMMENT ON COLUMN agent_licenses.license_status IS 'License status: ACTIVE, EXPIRED, or RENEWED (BR-AGT-PRF-013)';
COMMENT ON COLUMN agent_licenses.licentiate_exam_passed IS 'Flag indicating if licentiate exam passed (BR-AGT-PRF-012)';
COMMENT ON COLUMN agent_licenses.is_primary IS 'Primary license flag (one primary per agent)';

-- ----------------------------------------------------------------------------
-- Table: agent_license_reminders Comments
-- ----------------------------------------------------------------------------
COMMENT ON TABLE agent_license_reminders IS 'E-07: Agent License Reminder Log Entity - License renewal reminder tracking (BR-AGT-PRF-014)';

COMMENT ON COLUMN agent_license_reminders.reminder_type IS 'Reminder type: 30_DAYS, 15_DAYS, 7_DAYS, or EXPIRY_DAY (BR-AGT-PRF-014)';
COMMENT ON COLUMN agent_license_reminders.reminder_date IS 'Reminder scheduled date';
COMMENT ON COLUMN agent_license_reminders.sent_status IS 'Reminder sent status: PENDING, SENT, or FAILED';
COMMENT ON COLUMN agent_license_reminders.email_sent IS 'Email sent flag';
COMMENT ON COLUMN agent_license_reminders.sms_sent IS 'SMS sent flag';

-- ----------------------------------------------------------------------------
-- Table: agent_audit_logs Comments
-- ----------------------------------------------------------------------------
COMMENT ON TABLE agent_audit_logs IS 'E-08: Agent Audit Log Entity - Complete audit trail for all changes (BR-AGT-PRF-005, BR-AGT-PRF-006)';

COMMENT ON COLUMN agent_audit_logs.action_type IS 'Action type: CREATE, UPDATE, DELETE, STATUS_CHANGE, etc.';
COMMENT ON COLUMN agent_audit_logs.field_name IS 'Field that was changed';
COMMENT ON COLUMN agent_audit_logs.old_value IS 'Previous value before change';
COMMENT ON COLUMN agent_audit_logs.new_value IS 'New value after change';
COMMENT ON COLUMN agent_audit_logs.action_reason IS 'Reason for the change';
COMMENT ON COLUMN agent_audit_logs.performed_by IS 'User who performed the action (BR-AGT-PRF-005)';
COMMENT ON COLUMN agent_audit_logs.performed_at IS 'Timestamp when action was performed';
COMMENT ON COLUMN agent_audit_logs.ip_address IS 'IP address of the user for security';

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================

-- Total Tables: 8
-- Total Indexes: 50+
-- Total Functions: 5
-- Total Triggers: 7
-- Total Views: 4
-- Total Constraints: 20+
-- Postgres Version: 16
-- Extensions: uuid-ossp, pgcrypto
