package store

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type TxStatus int

const (
	TxStatusPending TxStatus = 0
	TxStatusSuccess TxStatus = 1
	TxStatusFailed  TxStatus = 2
)

type TxHistoryEntry struct {
	ID          int64     `json:"id"`
	TxHash      string    `json:"txHash"`
	FromAddr    string    `json:"fromAddr"`
	ToAddr      string    `json:"toAddr"`
	Value       string    `json:"value"`
	GasLimit    uint64    `json:"gasLimit"`
	GasPrice    string    `json:"gasPrice"`
	Nonce       uint64    `json:"nonce"`
	Data        string    `json:"data"`
	Status      TxStatus  `json:"status"`
	BlockNumber   uint64    `json:"blockNumber"`
	Network       string    `json:"network"`
	TxType        string    `json:"txType"`
	ContractAddr  string    `json:"contractAddr"`
	FailureReason string    `json:"failureReason"`
	CreatedAt     time.Time `json:"createdAt"`
}

type TxHistoryStore struct {
	db *sql.DB
}

func NewTxHistoryStore(dbPath string) (*TxHistoryStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = createTxHistoryTable(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	log.Printf("📦 [TxHistoryStore] SQLite 存储初始化成功，路径: %s", dbPath)
	return &TxHistoryStore{db: db}, nil
}

func createTxHistoryTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS tx_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tx_hash TEXT UNIQUE NOT NULL,
		from_addr TEXT NOT NULL,
		to_addr TEXT,
		value TEXT NOT NULL,
		gas_limit INTEGER,
		gas_price TEXT,
		nonce INTEGER,
		data TEXT,
		status INTEGER DEFAULT 0,
		block_number INTEGER,
		network TEXT NOT NULL,
		tx_type TEXT,
		contract_addr TEXT,
		failure_reason TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_tx_hash ON tx_history(tx_hash);
	CREATE INDEX IF NOT EXISTS idx_from_addr ON tx_history(from_addr);
	CREATE INDEX IF NOT EXISTS idx_network ON tx_history(network);
	`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	// 为现有表添加缺失列（处理 schema 迁移）
	_, _ = db.Exec(`ALTER TABLE tx_history ADD COLUMN contract_addr TEXT`)
	_, _ = db.Exec(`ALTER TABLE tx_history ADD COLUMN failure_reason TEXT`)

	return nil
}

func (s *TxHistoryStore) Add(entry TxHistoryEntry) error {
	_, err := s.db.Exec(`
	INSERT INTO tx_history (
		tx_hash, from_addr, to_addr, value, gas_limit, gas_price,
		nonce, data, status, block_number, network, tx_type, contract_addr, failure_reason, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(tx_hash) DO UPDATE SET status = excluded.status, block_number = excluded.block_number, failure_reason = excluded.failure_reason`,
		entry.TxHash, entry.FromAddr, entry.ToAddr, entry.Value,
		entry.GasLimit, entry.GasPrice, entry.Nonce, entry.Data,
		entry.Status, entry.BlockNumber, entry.Network, entry.TxType, entry.ContractAddr, entry.FailureReason, entry.CreatedAt)

	if err != nil {
		log.Printf("❌ [TxHistoryStore] 添加交易历史失败: %v", err)
		return err
	}

	log.Printf("💾 [TxHistoryStore] 交易历史已保存: %s", entry.TxHash)
	return nil
}

func (s *TxHistoryStore) GetByHash(txHash string) (*TxHistoryEntry, error) {
	row := s.db.QueryRow(`
	SELECT id, tx_hash, from_addr, to_addr, value, gas_limit, gas_price,
	       nonce, data, status, block_number, network, tx_type, contract_addr, failure_reason, created_at
	FROM tx_history WHERE tx_hash = ?`, txHash)

	var entry TxHistoryEntry
	err := row.Scan(
		&entry.ID, &entry.TxHash, &entry.FromAddr, &entry.ToAddr,
		&entry.Value, &entry.GasLimit, &entry.GasPrice, &entry.Nonce,
		&entry.Data, &entry.Status, &entry.BlockNumber, &entry.Network,
		&entry.TxType, &entry.ContractAddr, &entry.FailureReason, &entry.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		log.Printf("❌ [TxHistoryStore] 查询交易历史失败: %v", err)
		return nil, err
	}

	return &entry, nil
}

func (s *TxHistoryStore) List(limit, offset int) ([]TxHistoryEntry, error) {
	rows, err := s.db.Query(`
	SELECT id, tx_hash, from_addr, to_addr, value, gas_limit, gas_price,
	       nonce, data, status, block_number, network, tx_type, contract_addr, failure_reason, created_at
	FROM tx_history ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		log.Printf("❌ [TxHistoryStore] 查询交易列表失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var entries []TxHistoryEntry
	for rows.Next() {
		var entry TxHistoryEntry
		err := rows.Scan(
			&entry.ID, &entry.TxHash, &entry.FromAddr, &entry.ToAddr,
			&entry.Value, &entry.GasLimit, &entry.GasPrice, &entry.Nonce,
			&entry.Data, &entry.Status, &entry.BlockNumber, &entry.Network,
			&entry.TxType, &entry.ContractAddr, &entry.FailureReason, &entry.CreatedAt)
		if err != nil {
			log.Printf("❌ [TxHistoryStore] 扫描交易记录失败: %v", err)
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		log.Printf("❌ [TxHistoryStore] 遍历交易记录失败: %v", err)
		return nil, err
	}

	return entries, nil
}

func (s *TxHistoryStore) ListByType(txType string, limit, offset int) ([]TxHistoryEntry, error) {
	rows, err := s.db.Query(`
	SELECT id, tx_hash, from_addr, to_addr, value, gas_limit, gas_price,
	       nonce, data, status, block_number, network, tx_type, contract_addr, failure_reason, created_at
	FROM tx_history WHERE tx_type = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, txType, limit, offset)
	if err != nil {
		log.Printf("❌ [TxHistoryStore] 按类型查询交易列表失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var entries []TxHistoryEntry
	for rows.Next() {
		var entry TxHistoryEntry
		err := rows.Scan(
			&entry.ID, &entry.TxHash, &entry.FromAddr, &entry.ToAddr,
			&entry.Value, &entry.GasLimit, &entry.GasPrice, &entry.Nonce,
			&entry.Data, &entry.Status, &entry.BlockNumber, &entry.Network,
			&entry.TxType, &entry.ContractAddr, &entry.FailureReason, &entry.CreatedAt)
		if err != nil {
			log.Printf("❌ [TxHistoryStore] 按类型扫描交易记录失败: %v", err)
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		log.Printf("❌ [TxHistoryStore] 按类型遍历交易记录失败: %v", err)
		return nil, err
	}

	return entries, nil
}

func (s *TxHistoryStore) ListByTypeAndAddress(txType, address string, limit, offset int) ([]TxHistoryEntry, error) {
	rows, err := s.db.Query(`
	SELECT id, tx_hash, from_addr, to_addr, value, gas_limit, gas_price,
	       nonce, data, status, block_number, network, tx_type, contract_addr, failure_reason, created_at
	FROM tx_history WHERE tx_type = ? AND (LOWER(from_addr) = ? OR LOWER(to_addr) = ? OR LOWER(contract_addr) = ?)
	ORDER BY created_at DESC LIMIT ? OFFSET ?`, txType, address, address, address, limit, offset)
	if err != nil {
		log.Printf("❌ [TxHistoryStore] 按类型和地址查询交易列表失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var entries []TxHistoryEntry
	for rows.Next() {
		var entry TxHistoryEntry
		err := rows.Scan(
			&entry.ID, &entry.TxHash, &entry.FromAddr, &entry.ToAddr,
			&entry.Value, &entry.GasLimit, &entry.GasPrice, &entry.Nonce,
			&entry.Data, &entry.Status, &entry.BlockNumber, &entry.Network,
			&entry.TxType, &entry.ContractAddr, &entry.FailureReason, &entry.CreatedAt)
		if err != nil {
			log.Printf("❌ [TxHistoryStore] 按类型和地址扫描交易记录失败: %v", err)
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		log.Printf("❌ [TxHistoryStore] 按类型和地址遍历交易记录失败: %v", err)
		return nil, err
	}

	return entries, nil
}

func (s *TxHistoryStore) UpdateStatus(txHash string, status TxStatus, blockNumber uint64) error {
	result, err := s.db.Exec(`
	UPDATE tx_history SET status = ?, block_number = ? WHERE tx_hash = ?`,
		status, blockNumber, txHash)
	if err != nil {
		log.Printf("❌ [TxHistoryStore] 更新交易状态失败: %v", err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		log.Printf("⚠️  [TxHistoryStore] 更新状态时未找到交易: %s", txHash)
	}

	log.Printf("📝 [TxHistoryStore] 交易状态已更新: %s -> %d", txHash, status)
	return nil
}

func (s *TxHistoryStore) UpdateStatusWithReason(txHash string, status TxStatus, blockNumber uint64, failureReason string) error {
	result, err := s.db.Exec(`
	UPDATE tx_history SET status = ?, block_number = ?, failure_reason = ? WHERE tx_hash = ?`,
		status, blockNumber, failureReason, txHash)
	if err != nil {
		log.Printf("❌ [TxHistoryStore] 更新交易状态失败: %v", err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		log.Printf("⚠️  [TxHistoryStore] 更新状态时未找到交易: %s", txHash)
	}

	log.Printf("📝 [TxHistoryStore] 交易状态已更新: %s -> %d, 失败原因: %s", txHash, status, failureReason)
	return nil
}

func (s *TxHistoryStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM tx_history").Scan(&count)
	if err != nil {
		log.Printf("❌ [TxHistoryStore] 统计交易总数失败: %v", err)
		return 0, err
	}
	return count, nil
}

func (s *TxHistoryStore) CountByType(txType string) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM tx_history WHERE tx_type = ?", txType).Scan(&count)
	if err != nil {
		log.Printf("❌ [TxHistoryStore] 按类型统计失败: %v", err)
		return 0, err
	}
	return count, nil
}

func (s *TxHistoryStore) CountByTypeAndAddress(txType, address string) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM tx_history WHERE tx_type = ? AND (LOWER(from_addr) = ? OR LOWER(to_addr) = ? OR LOWER(contract_addr) = ?)", txType, address, address, address).Scan(&count)
	if err != nil {
		log.Printf("❌ [TxHistoryStore] 按类型和地址统计失败: %v", err)
		return 0, err
	}
	return count, nil
}

func (s *TxHistoryStore) ListByAddress(address string, limit, offset int) ([]TxHistoryEntry, error) {
	rows, err := s.db.Query(`
	SELECT id, tx_hash, from_addr, to_addr, value, gas_limit, gas_price,
	       nonce, data, status, block_number, network, tx_type, contract_addr, failure_reason, created_at
	FROM tx_history WHERE LOWER(from_addr) = ? OR LOWER(to_addr) = ? OR LOWER(contract_addr) = ?
	ORDER BY created_at DESC LIMIT ? OFFSET ?`, address, address, address, limit, offset)
	if err != nil {
		log.Printf("❌ [TxHistoryStore] 按地址查询交易列表失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var entries []TxHistoryEntry
	for rows.Next() {
		var entry TxHistoryEntry
		err := rows.Scan(
			&entry.ID, &entry.TxHash, &entry.FromAddr, &entry.ToAddr,
			&entry.Value, &entry.GasLimit, &entry.GasPrice, &entry.Nonce,
			&entry.Data, &entry.Status, &entry.BlockNumber, &entry.Network,
			&entry.TxType, &entry.ContractAddr, &entry.FailureReason, &entry.CreatedAt)
		if err != nil {
			log.Printf("❌ [TxHistoryStore] 按地址扫描交易记录失败: %v", err)
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		log.Printf("❌ [TxHistoryStore] 按地址遍历交易记录失败: %v", err)
		return nil, err
	}

	return entries, nil
}

func (s *TxHistoryStore) CountByAddress(address string) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM tx_history WHERE LOWER(from_addr) = ? OR LOWER(to_addr) = ? OR LOWER(contract_addr) = ?", address, address, address).Scan(&count)
	if err != nil {
		log.Printf("❌ [TxHistoryStore] 按地址统计失败: %v", err)
		return 0, err
	}
	return count, nil
}

func (s *TxHistoryStore) Close() error {
	return s.db.Close()
}

// ListByContractAddr 查询 contract_addr 为 NULL 或匹配 address 的记录（用于回填）
func (s *TxHistoryStore) ListByContractAddr(address string, limit, offset int) ([]TxHistoryEntry, error) {
	var rows *sql.Rows
	var err error
	if address == "" {
		// 回填：查询所有 contract_addr 为空的 erc20_transfer 记录
		rows, err = s.db.Query(`
		SELECT id, tx_hash, from_addr, to_addr, value, gas_limit, gas_price,
		       nonce, data, status, block_number, network, tx_type, contract_addr, created_at
		FROM tx_history WHERE tx_type = 'erc20_transfer' AND (contract_addr = '' OR contract_addr IS NULL)
		ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	} else {
		rows, err = s.db.Query(`
		SELECT id, tx_hash, from_addr, to_addr, value, gas_limit, gas_price,
		       nonce, data, status, block_number, network, tx_type, contract_addr, created_at
		FROM tx_history WHERE LOWER(contract_addr) = ?
		ORDER BY created_at DESC LIMIT ? OFFSET ?`, address, limit, offset)
	}
	if err != nil {
		log.Printf("❌ [TxHistoryStore] ListByContractAddr 失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var entries []TxHistoryEntry
	for rows.Next() {
		var entry TxHistoryEntry
		err := rows.Scan(
			&entry.ID, &entry.TxHash, &entry.FromAddr, &entry.ToAddr,
			&entry.Value, &entry.GasLimit, &entry.GasPrice, &entry.Nonce,
			&entry.Data, &entry.Status, &entry.BlockNumber, &entry.Network,
			&entry.TxType, &entry.ContractAddr, &entry.CreatedAt)
		if err != nil {
			log.Printf("❌ [TxHistoryStore] 扫描 contract_addr 记录失败: %v", err)
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

// UpdateContractAddr 根据交易哈希更新 contract_addr
func (s *TxHistoryStore) UpdateContractAddr(txHash, contractAddr string) error {
	_, err := s.db.Exec("UPDATE tx_history SET contract_addr = ? WHERE tx_hash = ?", contractAddr, txHash)
	if err != nil {
		return err
	}
	return nil
}

func (s *TxHistoryStore) CountByContractAddr(address string) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM tx_history WHERE LOWER(contract_addr) = ?", address).Scan(&count)
	if err != nil {
		log.Printf("❌ [TxHistoryStore] 按 contract_addr 统计失败: %v", err)
		return 0, err
	}
	return count, nil
}
