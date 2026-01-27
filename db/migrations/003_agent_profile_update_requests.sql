-- Migration: Agent Profile Update Requests Table
-- Phase 6.1: Approval Workflow for Critical Field Updates
-- BR-AGT-PRF-005: Name Update with Audit Logging
-- BR-AGT-PRF-006: PAN Update with Validation

-- Table: agent_profile_update_requests
-- Stores profile update requests that require approval (critical fields)
CREATE TABLE IF NOT EXISTS agent_profile_update_requests (
    -- Primary Key
    request_id TEXT PRIMARY KEY,

    -- Agent Information
    agent_id TEXT NOT NULL REFERENCES agent_profiles(agent_id) ON DELETE CASCADE,

    -- Request Details
    section TEXT NOT NULL CHECK (section IN ('personal_info', 'address', 'contact', 'email')),
    field_updates JSONB NOT NULL, -- {"field_name": "new_value"}
    reason TEXT,
    requested_by TEXT NOT NULL,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Approval Status
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
    approved_by TEXT,
    approved_at TIMESTAMPTZ,
    rejected_by TEXT,
    rejected_at TIMESTAMPTZ,
    comments TEXT,

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT chk_approval_logic CHECK (
        (status = 'PENDING' AND approved_by IS NULL AND rejected_by IS NULL)
        OR (status = 'APPROVED' AND approved_by IS NOT NULL AND approved_at IS NOT NULL)
        OR (status = 'REJECTED' AND rejected_by IS NOT NULL AND rejected_at IS NOT NULL)
    )
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_profile_update_requests_agent_id
    ON agent_profile_update_requests(agent_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_profile_update_requests_status
    ON agent_profile_update_requests(status) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_profile_update_requests_requested_at
    ON agent_profile_update_requests(requested_at DESC) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_profile_update_requests_agent_status
    ON agent_profile_update_requests(agent_id, status) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE agent_profile_update_requests IS 'Stores profile update requests for critical fields that require approval';
COMMENT ON COLUMN agent_profile_update_requests.field_updates IS 'JSON object containing field names and new values: {"field_name": "new_value"}';
COMMENT ON COLUMN agent_profile_update_requests.section IS 'Profile section being updated: personal_info, address, contact, email';
COMMENT ON COLUMN agent_profile_update_requests.status IS 'Request status: PENDING (awaiting approval), APPROVED (changes applied), REJECTED (changes discarded)';
COMMENT ON CONSTRAINT chk_approval_logic ON agent_profile_update_requests IS 'Ensures approval/rejection fields are set correctly based on status';
