-- Migration: 007_agent_search_dashboard_schema.sql
-- Phase 9: Search & Dashboard APIs
-- Purpose: Add notifications table and search optimization indexes
-- Date: 2026-01-27

-- ============================================================================
-- AGENT NOTIFICATIONS TABLE
-- ============================================================================
-- Purpose: Store email, SMS, and internal notification history for agents
-- FR-AGT-PRF-021: Self-Service Update notifications
-- AGT-077: Agent Notification History API

CREATE TABLE IF NOT EXISTS agent_notifications (
    notification_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL REFERENCES agent_profiles(agent_id) ON DELETE CASCADE,
    notification_type VARCHAR(50) NOT NULL, -- EMAIL, SMS, INTERNAL
    template VARCHAR(100) NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    subject TEXT,
    message TEXT,
    sent_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMP WITH TIME ZONE,
    read_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'SENT', -- SENT, DELIVERED, READ, FAILED
    failure_reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_notification_agent FOREIGN KEY (agent_id) REFERENCES agent_profiles(agent_id),
    CONSTRAINT chk_notification_type CHECK (notification_type IN ('EMAIL', 'SMS', 'INTERNAL')),
    CONSTRAINT chk_notification_status CHECK (status IN ('SENT', 'DELIVERED', 'READ', 'FAILED'))
);

-- ============================================================================
-- INDEXES FOR AGENT NOTIFICATIONS
-- ============================================================================

-- Primary lookup by agent_id
CREATE INDEX idx_notifications_agent_id ON agent_notifications(agent_id);

-- Query by sent date for timeline views
CREATE INDEX idx_notifications_sent_at ON agent_notifications(sent_at DESC);

-- Filter by notification type
CREATE INDEX idx_notifications_type ON agent_notifications(notification_type);

-- Filter by status for retry logic
CREATE INDEX idx_notifications_status ON agent_notifications(status);

-- Composite index for common query pattern (agent + type + date)
CREATE INDEX idx_notifications_agent_type_date ON agent_notifications(agent_id, notification_type, sent_at DESC);

-- ============================================================================
-- INDEXES FOR MULTI-CRITERIA SEARCH (AGT-022)
-- ============================================================================

-- Composite index for common search patterns
-- Supports queries filtering by status, agent_type, and office
CREATE INDEX IF NOT EXISTS idx_agent_search_composite ON agent_profiles(status, agent_type, office_code);

-- PAN number exact match (frequently used for search)
CREATE INDEX IF NOT EXISTS idx_agent_pan ON agent_profiles(pan_number) WHERE pan_number IS NOT NULL;

-- Name search optimization using trigram indexes (PostgreSQL extension)
-- Enables fast ILIKE queries for partial name matches
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_agent_name_trgm ON agent_profiles USING gin ((first_name || ' ' || last_name) gin_trgm_ops);

-- Employee ID lookup
CREATE INDEX IF NOT EXISTS idx_agent_employee_id ON agent_profiles(employee_id) WHERE employee_id IS NOT NULL;

-- ============================================================================
-- INDEXES FOR AGENT CONTACTS (Mobile Number Search)
-- ============================================================================

-- Mobile number exact match for search (AGT-022)
CREATE INDEX IF NOT EXISTS idx_contact_mobile ON agent_contacts(mobile_number) WHERE mobile_number IS NOT NULL;

-- Primary contact optimization
CREATE INDEX IF NOT EXISTS idx_contact_primary ON agent_contacts(agent_id) WHERE is_primary = true;

-- ============================================================================
-- INDEXES FOR AGENT EMAILS (Email Search)
-- ============================================================================

-- Email exact match for search
CREATE INDEX IF NOT EXISTS idx_email_address ON agent_emails(email_id) WHERE email_id IS NOT NULL;

-- Primary email optimization
CREATE INDEX IF NOT EXISTS idx_email_primary ON agent_emails(agent_id) WHERE is_primary = true;

-- ============================================================================
-- INDEXES FOR AUDIT LOGS (AGT-028, AGT-076)
-- ============================================================================

-- Audit history queries (date range filtering)
CREATE INDEX IF NOT EXISTS idx_audit_log_performed_at ON agent_audit_logs(performed_at DESC);

-- Composite index for agent audit queries
CREATE INDEX IF NOT EXISTS idx_audit_log_agent_date ON agent_audit_logs(agent_id, performed_at DESC);

-- Filter by action type for timeline views
CREATE INDEX IF NOT EXISTS idx_audit_log_action_type ON agent_audit_logs(action_type);

-- ============================================================================
-- INDEXES FOR LICENSE EXPIRY (Dashboard Warnings - AGT-068)
-- ============================================================================

-- Find licenses expiring soon for dashboard warnings
CREATE INDEX IF NOT EXISTS idx_license_expiry ON agent_licenses(expiry_date) WHERE status = 'ACTIVE';

-- Agent's active licenses
CREATE INDEX IF NOT EXISTS idx_license_agent_status ON agent_licenses(agent_id, status);

-- ============================================================================
-- OPTIMIZATION FOR HIERARCHY QUERIES (AGT-073)
-- ============================================================================

-- Advisor coordinator linkage for hierarchy traversal
CREATE INDEX IF NOT EXISTS idx_profile_coordinator ON agent_profiles(advisor_coordinator_id) WHERE advisor_coordinator_id IS NOT NULL;

-- Circle and division hierarchy
CREATE INDEX IF NOT EXISTS idx_profile_circle ON agent_profiles(circle_id) WHERE circle_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_profile_division ON agent_profiles(division_id) WHERE division_id IS NOT NULL;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE agent_notifications IS 'Phase 9: Stores email, SMS, and internal notification history for agents';
COMMENT ON INDEX idx_agent_search_composite IS 'Phase 9: Optimizes multi-criteria agent search by status, type, and office';
COMMENT ON INDEX idx_agent_name_trgm IS 'Phase 9: Enables fast partial name search using PostgreSQL trigram indexes';
COMMENT ON INDEX idx_audit_log_agent_date IS 'Phase 9: Optimizes audit history queries with date range filtering';

-- End of migration 007
