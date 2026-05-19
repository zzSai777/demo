CREATE TABLE IF NOT EXISTS control_configs (
  config_key VARCHAR(191) NOT NULL PRIMARY KEY,
  config_value TEXT NOT NULL,
  scope VARCHAR(64) NOT NULL DEFAULT 'global',
  version BIGINT NOT NULL DEFAULT 1,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS control_ab_tests (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(191) NOT NULL,
  feature_key VARCHAR(191) NOT NULL,
  variants_json JSON NOT NULL,
  traffic_percent INT NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_ab_feature_status (feature_key, status)
);

CREATE TABLE IF NOT EXISTS control_rollouts (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  feature_key VARCHAR(191) NOT NULL,
  target_percent INT NOT NULL,
  strategy VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_rollout_feature_status (feature_key, status)
);

CREATE TABLE IF NOT EXISTS control_services (
  service_name VARCHAR(128) NOT NULL PRIMARY KEY,
  status VARCHAR(32) NOT NULL,
  version VARCHAR(128) NULL,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS control_update_plans (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_name VARCHAR(128) NOT NULL,
  version VARCHAR(128) NOT NULL,
  strategy VARCHAR(64) NOT NULL DEFAULT 'rolling',
  status VARCHAR(32) NOT NULL DEFAULT 'planned',
  requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_update_service_status (service_name, status)
);

CREATE TABLE IF NOT EXISTS control_service_versions (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_name VARCHAR(128) NOT NULL,
  version VARCHAR(128) NOT NULL,
  artifact VARCHAR(512) NOT NULL,
  checksum VARCHAR(191) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'available',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_service_version (service_name, version),
  KEY idx_version_service_status (service_name, status)
);

CREATE TABLE IF NOT EXISTS control_releases (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_name VARCHAR(128) NOT NULL,
  version VARCHAR(128) NOT NULL,
  previous_version VARCHAR(128) NULL,
  strategy VARCHAR(64) NOT NULL DEFAULT 'rolling',
  status VARCHAR(32) NOT NULL DEFAULT 'released',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_release_service_created (service_name, created_at),
  KEY idx_release_service_status (service_name, status)
);

CREATE TABLE IF NOT EXISTS control_rollbacks (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  service_name VARCHAR(128) NOT NULL,
  target_version VARCHAR(128) NOT NULL,
  from_version VARCHAR(128) NULL,
  reason VARCHAR(512) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'rolled_back',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_rollback_service_created (service_name, created_at)
);

CREATE TABLE IF NOT EXISTS control_node_loads (
  node_id VARCHAR(128) NOT NULL PRIMARY KEY,
  service_name VARCHAR(128) NOT NULL,
  status VARCHAR(32) NOT NULL,
  cpu_percent INT NOT NULL DEFAULT 0,
  mem_percent INT NOT NULL DEFAULT 0,
  rooms INT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_node_service_status (service_name, status)
);
