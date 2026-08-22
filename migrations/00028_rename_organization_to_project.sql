-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'organizations'
    ) THEN
        ALTER TABLE organizations RENAME TO projects;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'user_organizations'
    ) THEN
        ALTER TABLE user_organizations RENAME TO user_projects;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'user_projects' AND column_name = 'organization_id'
    ) THEN
        ALTER TABLE user_projects RENAME COLUMN organization_id TO project_id;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'active_organization_id'
    ) THEN
        ALTER TABLE users RENAME COLUMN active_organization_id TO active_project_id;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'workflows' AND column_name = 'organization_id'
    ) THEN
        ALTER TABLE workflows RENAME COLUMN organization_id TO project_id;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'endpoints' AND column_name = 'organization_id'
    ) THEN
        ALTER TABLE endpoints RENAME COLUMN organization_id TO project_id;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'steps' AND column_name = 'organization_id'
    ) THEN
        ALTER TABLE steps RENAME COLUMN organization_id TO project_id;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'connections' AND column_name = 'organization_id'
    ) THEN
        ALTER TABLE connections RENAME COLUMN organization_id TO project_id;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'step_runs' AND column_name = 'organization_id'
    ) THEN
        ALTER TABLE step_runs RENAME COLUMN organization_id TO project_id;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'quotas' AND column_name = 'max_organization_members'
    ) THEN
        ALTER TABLE quotas RENAME COLUMN max_organization_members TO max_project_members;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'quotas' AND column_name = 'max_project_members'
    ) THEN
        ALTER TABLE quotas RENAME COLUMN max_project_members TO max_organization_members;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'step_runs' AND column_name = 'project_id'
    ) THEN
        ALTER TABLE step_runs RENAME COLUMN project_id TO organization_id;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'connections' AND column_name = 'project_id'
    ) THEN
        ALTER TABLE connections RENAME COLUMN project_id TO organization_id;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'steps' AND column_name = 'project_id'
    ) THEN
        ALTER TABLE steps RENAME COLUMN project_id TO organization_id;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'endpoints' AND column_name = 'project_id'
    ) THEN
        ALTER TABLE endpoints RENAME COLUMN project_id TO organization_id;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'workflows' AND column_name = 'project_id'
    ) THEN
        ALTER TABLE workflows RENAME COLUMN project_id TO organization_id;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'active_project_id'
    ) THEN
        ALTER TABLE users RENAME COLUMN active_project_id TO active_organization_id;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'user_projects' AND column_name = 'project_id'
    ) THEN
        ALTER TABLE user_projects RENAME COLUMN project_id TO organization_id;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'user_projects'
    ) THEN
        ALTER TABLE user_projects RENAME TO user_organizations;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'projects'
    ) THEN
        ALTER TABLE projects RENAME TO organizations;
    END IF;
END $$;
-- +goose StatementEnd
