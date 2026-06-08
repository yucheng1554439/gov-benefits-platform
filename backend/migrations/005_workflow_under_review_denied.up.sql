INSERT INTO workflow_transitions (agency_id, from_status, to_status, required_role)
SELECT '22222222-2222-2222-2222-222222222201', 'under_review', 'denied', 'case_worker'
WHERE NOT EXISTS (
    SELECT 1 FROM workflow_transitions
    WHERE agency_id = '22222222-2222-2222-2222-222222222201'
      AND from_status = 'under_review'
      AND to_status = 'denied'
);
