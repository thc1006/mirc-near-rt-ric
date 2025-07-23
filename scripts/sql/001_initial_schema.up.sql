
CREATE TABLE IF NOT EXISTS policy_types (
    policy_type_id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    policy_schema JSONB,
    create_schema JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS policy_instances (
    policy_id VARCHAR(255) PRIMARY KEY,
    policy_type_id VARCHAR(255) NOT NULL REFERENCES policy_types(policy_type_id) ON DELETE CASCADE,
    policy_data JSONB,
    status VARCHAR(50) NOT NULL,
    status_reason TEXT,
    target_near_rt_ric VARCHAR(255),
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_updated_by VARCHAR(255),
    last_updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    enforced_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS ml_models (
    model_id VARCHAR(255) PRIMARY KEY,
    model_name VARCHAR(255) NOT NULL,
    model_version VARCHAR(50),
    model_type VARCHAR(50),
    description TEXT,
    model_data BYTEA,
    model_url TEXT,
    metadata JSONB,
    status VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deployed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS enrichment_jobs (
    ei_job_id VARCHAR(255) PRIMARY KEY,
    ei_type_id VARCHAR(255) NOT NULL,
    job_owner VARCHAR(255),
    job_data JSONB,
    target_uri TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50)
);
