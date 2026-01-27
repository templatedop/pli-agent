-- Migration: 008_agent_batch_webhook_schema.sql
-- Phase 10: Batch & Webhook APIs
-- Purpose: Add export configuration, export jobs, and webhook event tables
-- Date: 2026-01-27

-- ============================================================================
-- AGENT EXPORT CONFIGURATIONS TABLE
-- ============================================================================
-- Purpose: Store export configurations for agent profile exports
-- AGT-064: Configure Export Parameters
-- FR-AGT-PRF-025: Profile Export

CREATE TABLE IF NOT EXISTS agent_export_configs (
    export_config_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    export_name VARCHAR(255) NOT NULL,
    filters JSONB NOT NULL DEFAULT '{}',
    fields JSONB NOT NULL DEFAULT '[]',
    output_format VARCHAR(50) NOT NULL,
    estimated_records INTEGER NOT NULL DEFAULT 0,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT chk_output_format CHECK (output_format IN ('EXCEL', 'PDF', 'CSV'))
);

-- ============================================================================
-- AGENT EXPORT JOBS TABLE
-- ============================================================================
-- Purpose: Track asynchronous export job execution status
-- AGT-065: Execute Export Asynchronously
-- AGT-066: Get Export Status
-- WF-AGT-PRF-012: Profile Export Workflow

CREATE TABLE IF NOT EXISTS agent_export_jobs (
    export_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    export_config_id UUID NOT NULL REFERENCES agent_export_configs(export_config_id) ON DELETE CASCADE,
    requested_by VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'IN_PROGRESS',
    progress_percentage INTEGER NOT NULL DEFAULT 0,
    records_processed INTEGER DEFAULT 0,
    total_records INTEGER DEFAULT 0,
    file_url TEXT,
    file_size_bytes BIGINT,
    workflow_id TEXT,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    metadata JSONB,
    CONSTRAINT fk_export_config FOREIGN KEY (export_config_id) REFERENCES agent_export_configs(export_config_id),
    CONSTRAINT chk_export_status CHECK (status IN ('IN_PROGRESS', 'COMPLETED', 'FAILED', 'CANCELLED')),
    CONSTRAINT chk_progress CHECK (progress_percentage BETWEEN 0 AND 100)
);

-- ============================================================================
-- HRMS WEBHOOK EVENTS TABLE
-- ============================================================================
-- Purpose: Log incoming HRMS webhook events for processing and debugging
-- AGT-078: HRMS Webhook Receiver
-- INT-AGT-001: HRMS System Integration

CREATE TABLE IF NOT EXISTS hrms_webhook_events (
    event_id UUID PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    employee_id VARCHAR(50) NOT NULL,
    employee_data JSONB NOT NULL,
    signature VARCHAR(512) NOT NULL,
    signature_valid BOOLEAN DEFAULT false,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'RECEIVED',
    processing_result JSONB,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT chk_event_type CHECK (event_type IN ('EMPLOYEE_CREATED', 'EMPLOYEE_UPDATED', 'EMPLOYEE_TRANSFERRED', 'EMPLOYEE_TERMINATED')),
    CONSTRAINT chk_webhook_status CHECK (status IN ('RECEIVED', 'PROCESSING', 'PROCESSED', 'FAILED', 'RETRYING'))
);

-- ============================================================================
-- BATCH OPERATION LOGS TABLE
-- ============================================================================
-- Purpose: Log batch operations for license deactivation and other bulk actions
-- AGT-038: Batch Deactivate Expired Licenses
-- WF-AGT-PRF-007: License Deactivation Workflow

CREATE TABLE IF NOT EXISTS agent_batch_operation_logs (
    batch_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operation_type VARCHAR(50) NOT NULL,
    batch_date DATE NOT NULL,
    workflow_id TEXT,
    total_agents INTEGER NOT NULL DEFAULT 0,
    agents_processed INTEGER NOT NULL DEFAULT 0,
    agents_succeeded INTEGER NOT NULL DEFAULT 0,
    agents_failed INTEGER NOT NULL DEFAULT 0,
    dry_run BOOLEAN DEFAULT false,
    agent_ids TEXT[], -- Array of agent IDs processed
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'IN_PROGRESS',
    error_summary TEXT,
    CONSTRAINT chk_operation_type CHECK (operation_type IN ('LICENSE_DEACTIVATION', 'STATUS_UPDATE', 'BULK_NOTIFICATION')),
    CONSTRAINT chk_batch_status CHECK (status IN ('IN_PROGRESS', 'COMPLETED', 'FAILED'))
);

-- ============================================================================
-- INDEXES FOR EXPORT CONFIGURATIONS
-- ============================================================================

-- Query by created_by for user's export history
CREATE INDEX idx_export_config_created_by ON agent_export_configs(created_by);

-- Query by creation date for recent exports
CREATE INDEX idx_export_config_created_at ON agent_export_configs(created_at DESC);

-- ============================================================================
-- INDEXES FOR EXPORT JOBS
-- ============================================================================

-- Primary lookup by status for active jobs
CREATE INDEX idx_export_jobs_status ON agent_export_jobs(status);

-- Query by requester for user's export history
CREATE INDEX idx_export_jobs_requested_by ON agent_export_jobs(requested_by);

-- Query by workflow ID for Temporal workflow tracking
CREATE INDEX idx_export_jobs_workflow_id ON agent_export_jobs(workflow_id) WHERE workflow_id IS NOT NULL;

-- Composite index for user's active jobs
CREATE INDEX idx_export_jobs_user_status ON agent_export_jobs(requested_by, status);

-- Query by date for cleanup of old exports
CREATE INDEX idx_export_jobs_completed_at ON agent_export_jobs(completed_at) WHERE completed_at IS NOT NULL;

-- ============================================================================
-- INDEXES FOR WEBHOOK EVENTS
-- ============================================================================

-- Primary lookup by employee ID
CREATE INDEX idx_webhook_events_employee_id ON hrms_webhook_events(employee_id);

-- Query by event type
CREATE INDEX idx_webhook_events_type ON hrms_webhook_events(event_type);

-- Query by status for processing queue
CREATE INDEX idx_webhook_events_status ON hrms_webhook_events(status);

-- Query by received date for event history
CREATE INDEX idx_webhook_events_received_at ON hrms_webhook_events(received_at DESC);

-- Query failed events for retry
CREATE INDEX idx_webhook_events_retry ON hrms_webhook_events(next_retry_at, status) WHERE status = 'RETRYING';

-- ============================================================================
-- INDEXES FOR BATCH OPERATION LOGS
-- ============================================================================

-- Query by operation type
CREATE INDEX idx_batch_logs_operation_type ON agent_batch_operation_logs(operation_type);

-- Query by batch date for daily operations
CREATE INDEX idx_batch_logs_batch_date ON agent_batch_operation_logs(batch_date DESC);

-- Query by workflow ID for Temporal tracking
CREATE INDEX idx_batch_logs_workflow_id ON agent_batch_operation_logs(workflow_id) WHERE workflow_id IS NOT NULL;

-- Query by status for monitoring
CREATE INDEX idx_batch_logs_status ON agent_batch_operation_logs(status);

-- ============================================================================
-- AUTOMATIC CLEANUP POLICIES (Optional - for production)
-- ============================================================================

-- Cleanup old completed export jobs (older than 30 days)
-- COMMENT: Enable this in production after testing
-- CREATE OR REPLACE FUNCTION cleanup_old_export_jobs()
-- RETURNS void AS $$
-- BEGIN
--     DELETE FROM agent_export_jobs
--     WHERE status = 'COMPLETED'
--       AND completed_at < NOW() - INTERVAL '30 days';
-- END;
-- $$ LANGUAGE plpgsql;

-- Cleanup old processed webhook events (older than 90 days)
-- COMMENT: Enable this in production after testing
-- CREATE OR REPLACE FUNCTION cleanup_old_webhook_events()
-- RETURNS void AS $$
-- BEGIN
--     DELETE FROM hrms_webhook_events
--     WHERE status = 'PROCESSED'
--       AND processed_at < NOW() - INTERVAL '90 days';
-- END;
-- $$ LANGUAGE plpgsql;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE agent_export_configs IS 'Phase 10: Stores export configuration templates for agent profile exports';
COMMENT ON TABLE agent_export_jobs IS 'Phase 10: Tracks asynchronous export job execution with Temporal workflow integration';
COMMENT ON TABLE hrms_webhook_events IS 'Phase 10: Logs HRMS webhook events for employee data synchronization';
COMMENT ON TABLE agent_batch_operation_logs IS 'Phase 10: Logs bulk operations like license deactivation and status updates';

COMMENT ON INDEX idx_export_jobs_status IS 'Phase 10: Optimizes queries for active/pending export jobs';
COMMENT ON INDEX idx_webhook_events_retry IS 'Phase 10: Optimizes retry queue for failed webhook events';
COMMENT ON INDEX idx_batch_logs_batch_date IS 'Phase 10: Optimizes daily batch operation queries';

-- End of migration 008
