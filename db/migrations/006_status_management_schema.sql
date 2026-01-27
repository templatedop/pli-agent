-- Phase 8: Status Management Schema
-- AGT-039 to AGT-041: Termination and Reinstatement
-- BR-AGT-PRF-016: Status Update with Reason
-- BR-AGT-PRF-017: Agent Termination Workflow

-- Add termination tracking fields to agent_profiles
ALTER TABLE agent_profiles
    ADD COLUMN IF NOT EXISTS termination_date DATE,
    ADD COLUMN IF NOT EXISTS termination_reason TEXT,
    ADD COLUMN IF NOT EXISTS termination_reason_code VARCHAR(50),
    ADD COLUMN IF NOT EXISTS terminated_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS commission_enabled BOOLEAN DEFAULT true,
    ADD COLUMN IF NOT EXISTS reinstatement_date DATE,
    ADD COLUMN IF NOT EXISTS reinstated_by VARCHAR(100),
    ADD COLUMN IF NOT EXISTS reinstatement_reason TEXT;

-- Create termination_reason_code enum constraint
ALTER TABLE agent_profiles
    ADD CONSTRAINT check_termination_reason_code
    CHECK (termination_reason_code IS NULL OR termination_reason_code IN (
        'RESIGNATION', 'MISCONDUCT', 'NON_PERFORMANCE', 'FRAUD', 'LICENSE_EXPIRED', 'OTHER'
    ));

-- Create agent_termination_records table
-- Stores complete termination history with generated documents
CREATE TABLE IF NOT EXISTS agent_termination_records (
    termination_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL REFERENCES agent_profiles(agent_id),

    -- Termination details
    termination_date DATE NOT NULL,
    effective_date DATE NOT NULL,
    termination_reason TEXT NOT NULL CHECK (char_length(termination_reason) >= 20),
    termination_reason_code VARCHAR(50) NOT NULL,
    terminated_by VARCHAR(100) NOT NULL,

    -- Workflow tracking
    workflow_id TEXT,
    workflow_status VARCHAR(50) DEFAULT 'COMPLETED',

    -- Actions performed
    status_updated BOOLEAN DEFAULT false,
    portal_disabled BOOLEAN DEFAULT false,
    commission_stopped BOOLEAN DEFAULT false,
    letter_generated BOOLEAN DEFAULT false,
    data_archived BOOLEAN DEFAULT false,
    notifications_sent BOOLEAN DEFAULT false,

    -- Generated documents
    termination_letter_url TEXT,
    termination_letter_generated_at TIMESTAMPTZ,

    -- Metadata
    metadata JSONB,

    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    version INT NOT NULL DEFAULT 1,

    CONSTRAINT check_term_reason_code CHECK (termination_reason_code IN (
        'RESIGNATION', 'MISCONDUCT', 'NON_PERFORMANCE', 'FRAUD', 'LICENSE_EXPIRED', 'OTHER'
    ))
);

-- Create agent_reinstatement_requests table
-- Stores reinstatement requests with approval workflow
CREATE TABLE IF NOT EXISTS agent_reinstatement_requests (
    reinstatement_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL REFERENCES agent_profiles(agent_id),

    -- Request details
    request_date DATE NOT NULL DEFAULT CURRENT_DATE,
    reinstatement_reason TEXT NOT NULL CHECK (char_length(reinstatement_reason) >= 10),
    requested_by VARCHAR(100) NOT NULL,

    -- Approval workflow
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    approved_by VARCHAR(100),
    approved_at TIMESTAMPTZ,
    rejected_by VARCHAR(100),
    rejected_at TIMESTAMPTZ,
    rejection_reason TEXT,

    -- Workflow tracking
    workflow_id TEXT,

    -- Conditions and terms
    reinstatement_conditions TEXT,
    probation_period_days INT,

    -- Metadata
    metadata JSONB,

    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    version INT NOT NULL DEFAULT 1,

    CONSTRAINT check_reinst_status CHECK (status IN (
        'PENDING', 'APPROVED', 'REJECTED', 'COMPLETED'
    ))
);

-- Create agent_data_archives table
-- Tracks archived agent data for 7-year retention (BR-AGT-PRF-017)
CREATE TABLE IF NOT EXISTS agent_data_archives (
    archive_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL,

    -- Archive details
    archive_date DATE NOT NULL DEFAULT CURRENT_DATE,
    archive_type VARCHAR(50) NOT NULL,
    retention_until DATE NOT NULL, -- 7 years from termination

    -- Archived data
    data_snapshot JSONB NOT NULL,
    data_checksum TEXT,

    -- Storage reference
    storage_location TEXT,
    storage_size_bytes BIGINT,

    -- Metadata
    archived_by VARCHAR(100) NOT NULL,
    metadata JSONB,

    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT check_archive_type CHECK (archive_type IN (
        'TERMINATION', 'REINSTATEMENT', 'PERIODIC', 'MANUAL'
    ))
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_term_records_agent_id ON agent_termination_records(agent_id);
CREATE INDEX IF NOT EXISTS idx_term_records_date ON agent_termination_records(termination_date);
CREATE INDEX IF NOT EXISTS idx_term_records_status ON agent_termination_records(workflow_status);

CREATE INDEX IF NOT EXISTS idx_reinst_requests_agent_id ON agent_reinstatement_requests(agent_id);
CREATE INDEX IF NOT EXISTS idx_reinst_requests_status ON agent_reinstatement_requests(status);
CREATE INDEX IF NOT EXISTS idx_reinst_requests_date ON agent_reinstatement_requests(request_date);

CREATE INDEX IF NOT EXISTS idx_archives_agent_id ON agent_data_archives(agent_id);
CREATE INDEX IF NOT EXISTS idx_archives_type ON agent_data_archives(archive_type);
CREATE INDEX IF NOT EXISTS idx_archives_retention ON agent_data_archives(retention_until);

-- Comments
COMMENT ON TABLE agent_termination_records IS 'Stores complete termination history with workflow tracking and generated documents';
COMMENT ON TABLE agent_reinstatement_requests IS 'Stores reinstatement requests with approval workflow';
COMMENT ON TABLE agent_data_archives IS 'Tracks archived agent data for 7-year retention (BR-AGT-PRF-017)';

COMMENT ON COLUMN agent_profiles.termination_reason IS 'Termination reason - min 20 characters (BR-AGT-PRF-016)';
COMMENT ON COLUMN agent_profiles.commission_enabled IS 'Flag to enable/disable commission processing';
COMMENT ON COLUMN agent_termination_records.workflow_id IS 'Temporal workflow ID for termination orchestration (WF-AGT-PRF-004)';
COMMENT ON COLUMN agent_reinstatement_requests.workflow_id IS 'Temporal workflow ID for reinstatement orchestration (WF-AGT-PRF-011)';
COMMENT ON COLUMN agent_data_archives.retention_until IS '7-year retention from termination date (BR-AGT-PRF-017)';
