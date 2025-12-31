-- Migration: 000004_projects (rollback)

DROP TRIGGER IF EXISTS projects_updated_at ON projects;
DROP TABLE IF EXISTS projects CASCADE;

