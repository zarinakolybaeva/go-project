CREATE TABLE IF NOT EXISTS tasks (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    description text NOT NULL,
    due_date timestamp(0) with time zone NOT NULL,
    priority text NOT NULL,
    status text NOT NULL,
    category text NOT NULL,
    user_id bigserial,
    version integer NOT NULL DEFAULT 1

    -- Define a foreign key constraint to link tasks with users (assuming users table)
--     FOREIGN KEY (user_id) REFERENCES users(id)
);