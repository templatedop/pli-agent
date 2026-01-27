-- Migration: Stored Function for Single-Hit Approval and Profile Update
-- Phase 6 Optimization: Reduce ApproveProfileUpdate from 2 database hits to 1
-- This function encapsulates: fetch request → update profile → create audits → approve request

CREATE OR REPLACE FUNCTION approve_request_and_update_profile(
    p_request_id TEXT,
    p_approved_by TEXT,
    p_comments TEXT
) RETURNS TABLE (
    profile_json JSONB,
    request_json JSONB
) AS $$
DECLARE
    v_agent_id TEXT;
    v_field_updates JSONB;
    v_field_key TEXT;
    v_field_value TEXT;
    v_old_value TEXT;
    v_update_sql TEXT;
    v_profile_record RECORD;
    v_old_profile_record RECORD;
    v_now TIMESTAMPTZ := NOW();
BEGIN
    -- Step 1: Fetch and validate update request
    SELECT agent_id, field_updates::jsonb
    INTO v_agent_id, v_field_updates
    FROM agent_profile_update_requests
    WHERE request_id = p_request_id
      AND status = 'PENDING'
      AND deleted_at IS NULL;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Update request not found or already processed: %', p_request_id;
    END IF;

    -- Step 2: Capture old profile values for audit
    EXECUTE format('SELECT * FROM agent_profiles WHERE agent_id = %L AND deleted_at IS NULL', v_agent_id)
    INTO v_old_profile_record;

    -- Step 3: Build dynamic UPDATE statement
    v_update_sql := 'UPDATE agent_profiles SET updated_at = $1, updated_by = $2, version = version + 1';

    -- Add dynamic fields from JSON
    FOR v_field_key IN SELECT jsonb_object_keys(v_field_updates)
    LOOP
        v_field_value := v_field_updates->>v_field_key;
        v_update_sql := v_update_sql || format(', %I = %L', v_field_key, v_field_value);
    END LOOP;

    v_update_sql := v_update_sql || format(' WHERE agent_id = %L AND deleted_at IS NULL RETURNING *', v_agent_id);

    -- Step 4: Execute update
    EXECUTE v_update_sql
    USING v_now, p_approved_by
    INTO v_profile_record;

    -- Step 5: Create audit logs for changed fields
    FOR v_field_key IN SELECT jsonb_object_keys(v_field_updates)
    LOOP
        -- Get old value (cast to text for comparison)
        EXECUTE format('SELECT ($1).%I::text', v_field_key)
        USING v_old_profile_record
        INTO v_old_value;

        v_field_value := v_field_updates->>v_field_key;

        -- Only insert audit if value actually changed
        IF v_old_value IS DISTINCT FROM v_field_value THEN
            INSERT INTO agent_audit_logs (
                agent_id, action_type, field_name, old_value, new_value,
                performed_by, performed_at
            ) VALUES (
                v_agent_id, 'UPDATE', v_field_key, v_old_value, v_field_value,
                p_approved_by, v_now
            );
        END IF;
    END LOOP;

    -- Step 6: Mark request as approved
    UPDATE agent_profile_update_requests
    SET
        status = 'APPROVED',
        approved_by = p_approved_by,
        approved_at = v_now,
        comments = p_comments,
        updated_at = v_now
    WHERE request_id = p_request_id;

    -- Step 7: Return results as JSON
    RETURN QUERY
    SELECT
        to_jsonb(v_profile_record) as profile_json,
        to_jsonb(req.*) as request_json
    FROM agent_profile_update_requests req
    WHERE req.request_id = p_request_id;

END;
$$ LANGUAGE plpgsql;

-- Usage example:
-- SELECT * FROM approve_request_and_update_profile('request-uuid', 'approver-id', 'Approved');

-- Comments
COMMENT ON FUNCTION approve_request_and_update_profile IS
'Approves profile update request and applies changes in single database call.
Returns updated profile and approved request as JSONB.
Optimized for minimal round trips: 1 function call replaces 2 separate operations.';
