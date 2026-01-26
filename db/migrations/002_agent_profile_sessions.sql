-- Agent Profile Creation Sessions Table
-- Stores workflow session state for profile creation process
-- WF-AGT-PRF-001: Profile Creation Workflow

CREATE TABLE IF NOT EXISTS agent_profile_sessions (
    -- Primary Key
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Agent Type
    agent_type TEXT NOT NULL CHECK (agent_type IN ('ADVISOR', 'ADVISOR_COORDINATOR', 'DEPARTMENTAL_EMPLOYEE', 'FIELD_OFFICER', 'DIRECT_AGENT', 'GDS')),

    -- Workflow State
    workflow_state TEXT NOT NULL DEFAULT 'INITIATED' CHECK (workflow_state IN (
        'INITIATED',
        'HRMS_FETCHING',
        'HRMS_FETCHED',
        'COORDINATOR_LINKING',
        'PROFILE_VALIDATION',
        'PROFILE_SUBMITTING',
        'COMPLETED',
        'CANCELLED',
        'EXPIRED'
    )),

    -- Current Step (for UI navigation)
    current_step TEXT,
    next_step TEXT,
    progress_percentage INTEGER DEFAULT 0 CHECK (progress_percentage >= 0 AND progress_percentage <= 100),

    -- Session Data (JSONB for flexibility)
    form_data JSONB DEFAULT '{}'::jsonb,
    validation_errors JSONB DEFAULT '[]'::jsonb,

    -- Temporal Workflow Integration
    temporal_workflow_id TEXT,
    temporal_run_id TEXT,

    -- Session Status
    status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'EXPIRED', 'COMPLETED', 'CANCELLED')),

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '24 hours'),
    completed_at TIMESTAMP WITH TIME ZONE,

    -- User Tracking
    initiated_by TEXT NOT NULL,
    last_updated_by TEXT,

    -- Metadata
    metadata JSONB DEFAULT '{}'::jsonb,

    -- Audit
    created_agent_id UUID REFERENCES agent_profiles(agent_id) ON DELETE SET NULL
);

-- Indexes for session queries
CREATE INDEX IF NOT EXISTS idx_sessions_status ON agent_profile_sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_workflow_state ON agent_profile_sessions(workflow_state);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON agent_profile_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_temporal_workflow ON agent_profile_sessions(temporal_workflow_id);
CREATE INDEX IF NOT EXISTS idx_sessions_initiated_by ON agent_profile_sessions(initiated_by);

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_session_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_session_updated_at
    BEFORE UPDATE ON agent_profile_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_session_updated_at();

-- Comments
COMMENT ON TABLE agent_profile_sessions IS 'Stores workflow session state for agent profile creation process';
COMMENT ON COLUMN agent_profile_sessions.session_id IS 'Unique session identifier';
COMMENT ON COLUMN agent_profile_sessions.agent_type IS 'Type of agent being created';
COMMENT ON COLUMN agent_profile_sessions.workflow_state IS 'Current workflow state';
COMMENT ON COLUMN agent_profile_sessions.form_data IS 'Session form data in JSON format';
COMMENT ON COLUMN agent_profile_sessions.temporal_workflow_id IS 'Temporal workflow ID for integration';
COMMENT ON COLUMN agent_profile_sessions.expires_at IS 'Session expiry time (default 24 hours)';
