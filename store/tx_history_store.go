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
	BlockNumber uint64    `json:"blockNumber"`
	Network     string    `json:"network"`
	TxType      string    `json:"txType"`
	CreatedAt   time.Time `json:"createdAt"`
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
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_tx_hash ON tx_history(tx_hash);
	CREATE INDEX IF NOT EXISTS idx_from_addr ON tx_history(from_addr);
	CREATE INDEX IF NOT EXISTS idx_network ON tx_history(network);
	`

	_, err := db.Exec(query)
	return err
}

func (s *TxHistoryStore) Add(entry TxHistoryEntry) error {
	_, err := s.db.Exec(`
	INSERT INTO tx_history (
		tx_hash, from_addr, to_addr, value, gas_limit, gas_price,
		nonce, data, status, block_number, network, tx_type, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(tx_hash) DO UPDATE SET status = excluded.status, block_number = excluded.block_number`,
		entry.TxHash, entry.FromAddr, entry.ToAddr, entry.Value,
		entry.GasLimit, entry.GasPrice, entry.Nonce, entry.Data,
		entry.Status, entry.BlockNumber, entry.Network, entry.TxType, entry.CreatedAt)

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
	       nonce, data, status, block_number, network, tx_type, created_at
	FROM tx_history WHERE tx_hash = ?`, txHash)

	var entry TxHistoryEntry
	err := row.Scan(
		&entry.ID, &entry.TxHash, &entry.FromAddr, &entry.ToAddr,
		&entry.Value, &entry.GasLimit, &entry.GasPrice, &entry.Nonce,
		&entry.Data, &entry.Status, &entry.BlockNumber, &entry.Network,
		&entry.TxType, &entry.CreatedAt)

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
	       nonce, data, status, block_number, network, tx_type, created_at
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
			&entry.TxType, &entry.CreatedAt)
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
	       nonce, data, status, block_number, network, tx_type, created_at
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
			&entry.TxType, &entry.CreatedAt)
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
	       nonce, data, status, block_number, network, tx_type, created_at
	FROM tx_history WHERE tx_type = ? AND (from_addr = ? OR to_addr = ?)
	ORDER BY created_at DESC LIMIT ? OFFSET ?`, txType, address, address, limit, offset)
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
			&entry.TxType, &entry.CreatedAt)
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
	err := s.db.QueryRow("SELECT COUNT(*) FROM tx_history WHERE tx_type = ? AND (from_addr = ? OR to_addr = ?)", txType, address, address).Scan(&count)
	if err != nil {
		log.Printf("❌ [TxHistoryStore] 按类型和地址统计失败: %v", err)
		return 0, err
	}
	return count, nil
}

func (s *TxHistoryStore) Close() error {
	return s.db.Close()
}
