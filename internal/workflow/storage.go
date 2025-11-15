package workflow

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"netcaptor/internal/ai"
	"netcaptor/internal/utils"
)

type WorkflowStorage struct {
	db *sql.DB
}

func NewWorkflowStorage() (*WorkflowStorage, error) {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户目录失败: %w", err)
	}

	// 创建数据目录
	dataDir := filepath.Join(homeDir, ".netcaptor")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 打开数据库
	dbPath := filepath.Join(dataDir, "workflow.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	storage := &WorkflowStorage{db: db}
	if err := storage.initTables(); err != nil {
		db.Close()
		return nil, err
	}

	utils.AppLog.Info(fmt.Sprintf("[Storage] 数据库已初始化: %s", dbPath))
	return storage, nil
}

func (s *WorkflowStorage) initTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS workflow_tasks (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		nodes_json TEXT NOT NULL,
		edges_json TEXT NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS ai_models (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		provider TEXT NOT NULL,
		name TEXT NOT NULL,
		api_key TEXT NOT NULL,
		base_url TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("创建表失败: %w", err)
	}
	return nil
}

func (s *WorkflowStorage) SaveTask(task WorkflowTask) error {
	nodesJSON, err := json.Marshal(task.Nodes)
	if err != nil {
		return fmt.Errorf("序列化节点失败: %w", err)
	}

	edgesJSON, err := json.Marshal(task.Edges)
	if err != nil {
		return fmt.Errorf("序列化边失败: %w", err)
	}

	query := `
	INSERT OR REPLACE INTO workflow_tasks (id, name, description, created_at, updated_at, nodes_json, edges_json)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err = s.db.Exec(query, task.ID, task.Name, task.Description, task.CreatedAt, task.UpdatedAt, string(nodesJSON), string(edgesJSON))
	if err != nil {
		return fmt.Errorf("保存任务失败: %w", err)
	}

	utils.AppLog.Info(fmt.Sprintf("[Storage] 任务已保存: %s (%s)", task.Name, task.ID))
	return nil
}

func (s *WorkflowStorage) GetTask(id string) (*WorkflowTask, error) {
	query := `SELECT id, name, description, created_at, updated_at, nodes_json, edges_json FROM workflow_tasks WHERE id = ?`

	var task WorkflowTask
	var nodesJSON, edgesJSON string

	err := s.db.QueryRow(query, id).Scan(&task.ID, &task.Name, &task.Description, &task.CreatedAt, &task.UpdatedAt, &nodesJSON, &edgesJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}

	if err := json.Unmarshal([]byte(nodesJSON), &task.Nodes); err != nil {
		return nil, fmt.Errorf("反序列化节点失败: %w", err)
	}

	if err := json.Unmarshal([]byte(edgesJSON), &task.Edges); err != nil {
		return nil, fmt.Errorf("反序列化边失败: %w", err)
	}

	return &task, nil
}

func (s *WorkflowStorage) GetAllTasks() ([]WorkflowTask, error) {
	query := `SELECT id, name, description, created_at, updated_at, nodes_json, edges_json FROM workflow_tasks ORDER BY updated_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询任务列表失败: %w", err)
	}
	defer rows.Close()

	tasks := []WorkflowTask{}
	for rows.Next() {
		var task WorkflowTask
		var nodesJSON, edgesJSON string

		if err := rows.Scan(&task.ID, &task.Name, &task.Description, &task.CreatedAt, &task.UpdatedAt, &nodesJSON, &edgesJSON); err != nil {
			return nil, fmt.Errorf("扫描任务失败: %w", err)
		}

		if err := json.Unmarshal([]byte(nodesJSON), &task.Nodes); err != nil {
			return nil, fmt.Errorf("反序列化节点失败: %w", err)
		}

		if err := json.Unmarshal([]byte(edgesJSON), &task.Edges); err != nil {
			return nil, fmt.Errorf("反序列化边失败: %w", err)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s *WorkflowStorage) DeleteTask(id string) error {
	query := `DELETE FROM workflow_tasks WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除任务失败: %w", err)
	}
	utils.AppLog.Info(fmt.Sprintf("[Storage] 任务已删除: %s", id))
	return nil
}

// SaveAIModels 保存AI模型配置
func (s *WorkflowStorage) SaveAIModels(models []ai.AIModel) error {
	cm := utils.GetConfigManager()
	err := cm.SetConfigJSON("ai_models", models)
	if err == nil {
		utils.AppLog.Info(fmt.Sprintf("[Storage] 已保存 %d 个AI模型配置", len(models)))
	}
	return err
}

// LoadAIModels 加载AI模型配置
func (s *WorkflowStorage) LoadAIModels() ([]ai.AIModel, error) {
	cm := utils.GetConfigManager()
	var models []ai.AIModel
	err := cm.GetConfigJSON("ai_models", &models)
	if err != nil {
		return []ai.AIModel{}, nil // 返回空数组而不是错误
	}
	utils.AppLog.Info(fmt.Sprintf("[Storage] 已加载 %d 个AI模型配置", len(models)))
	return models, nil
}

func (s *WorkflowStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
