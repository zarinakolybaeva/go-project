ALTER TABLE tasks ADD CONSTRAINT tasks_due_date_check CHECK (due_date > created_at);





