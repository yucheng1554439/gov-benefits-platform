DELETE FROM workflow_transitions
WHERE agency_id = '22222222-2222-2222-2222-222222222201'
  AND from_status = 'under_review'
  AND to_status = 'denied';
