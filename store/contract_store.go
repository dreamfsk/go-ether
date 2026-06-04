package store

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type ContractEntry struct {
	ID          int64     `json:"id"`
	Address     string    `json:"address"`
	Name        string    `json:"name"`
	Symbol      string    `json:"symbol"`
	Network     string    `json:"network"`
	Deployer    string    `json:"deployer"`
	TxHash      string    `json:"txHash"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ContractStore struct {
	db *sql.DB
}

func NewContractStore(dbPath string) (*ContractStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = createContractTable(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	log.Printf("📦 [ContractStore] SQLite 存储初始化成功，路径: %s", dbPath)
	return &ContractStore{db: db}, nil
}

func createContractTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS contracts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		address TEXT UNIQUE NOT NULL,
		name TEXT,
		symbol TEXT,
		network TEXT NOT NULL,
		deployer TEXT,
		tx_hash TEXT,
		is_active INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_contract_address ON contracts(address);
	CREATE INDEX IF NOT EXISTS idx_contract_network ON contracts(network);
	CREATE INDEX IF NOT EXISTS idx_contract_active ON contracts(is_active);
	`

	_, err := db.Exec(query)
	return err
}

func (s *ContractStore) Add(entry ContractEntry) error {
	_, err := s.db.Exec(`
	INSERT INTO contracts (address, name, symbol, network, deployer, tx_hash, is_active, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(address) DO UPDATE SET name = excluded.name, symbol = excluded.symbol, 
		network = excluded.network, deployer = excluded.deployer, tx_hash = excluded.tx_hash`,
		entry.Address, entry.Name, entry.Symbol, entry.Network, entry.Deployer, entry.TxHash, entry.IsActive, entry.CreatedAt)

	if err != nil {
		log.Printf("❌ [ContractStore] 添加合约失败: %v", err)
		return err
	}

	log.Printf("💾 [ContractStore] 合约已保存: %s", entry.Address)
	return nil
}

func (s *ContractStore) GetActive() (*ContractEntry, error) {
	row := s.db.QueryRow(`
	SELECT id, address, name, symbol, network, deployer, tx_hash, is_active, created_at
	FROM contracts WHERE is_active = 1 LIMIT 1`)

	var entry ContractEntry
	err := row.Scan(
		&entry.ID, &entry.Address, &entry.Name, &entry.Symbol,
		&entry.Network, &entry.Deployer, &entry.TxHash, &entry.IsActive, &entry.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		log.Printf("❌ [ContractStore] 查询活跃合约失败: %v", err)
		return nil, err
	}

	return &entry, nil
}

func (s *ContractStore) GetByAddress(address string) (*ContractEntry, error) {
	row := s.db.QueryRow(`
	SELECT id, address, name, symbol, network, deployer, tx_hash, is_active, created_at
	FROM contracts WHERE address = ?`, address)

	var entry ContractEntry
	err := row.Scan(
		&entry.ID, &entry.Address, &entry.Name, &entry.Symbol,
		&entry.Network, &entry.Deployer, &entry.TxHash, &entry.IsActive, &entry.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		log.Printf("❌ [ContractStore] 查询合约失败: %v", err)
		return nil, err
	}

	return &entry, nil
}

func (s *ContractStore) List() ([]ContractEntry, error) {
	rows, err := s.db.Query(`
	SELECT id, address, name, symbol, network, deployer, tx_hash, is_active, created_at
	FROM contracts ORDER BY created_at DESC`)
	if err != nil {
		log.Printf("❌ [ContractStore] 查询合约列表失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var entries []ContractEntry
	for rows.Next() {
		var entry ContractEntry
		err := rows.Scan(
			&entry.ID, &entry.Address, &entry.Name, &entry.Symbol,
			&entry.Network, &entry.Deployer, &entry.TxHash, &entry.IsActive, &entry.CreatedAt)
		if err != nil {
			log.Printf("❌ [ContractStore] 扫描合约记录失败: %v", err)
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *ContractStore) SetActive(address string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE contracts SET is_active = 0")
	if err != nil {
		tx.Rollback()
		return err
	}

	result, err := tx.Exec("UPDATE contracts SET is_active = 1 WHERE address = ?", address)
	if err != nil {
		tx.Rollback()
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		tx.Rollback()
		return fmt.Errorf("contract not found: %s", address)
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	log.Printf("✅ [ContractStore] 已切换到合约: %s", address)
	return nil
}

func (s *ContractStore) Close() error {
	return s.db.Close()
}
