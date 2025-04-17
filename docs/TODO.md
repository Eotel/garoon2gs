# 修正すべき問題点

## 1. 過去日付スキップ処理の改善

現在の実装では、過去の日付（現在の日付より前の日付）はスケジュール書き込みの対象から除外されています。これにより、例えば4月17日に実行した場合、5月1日から13日までの予定が書き込まれず、14日以降のみが書き込まれるという問題が発生しています。

### 問題のあるコード
```go
// 過去の日付はスキップ
if cellDate.Before(today) {
    log.Printf("Skipping past date: %s (before today: %s)", cellDate.Format("2006-01-02"), today.Format("2006-01-02"))
    continue
}
```

### 修正案
過去の日付でも、現在の月の日付か未来の月の日付であれば処理するように修正します：

```go
// 過去の日付でも、現在の月と同じか未来の月であれば処理する
isCurrentOrFutureMonth := cellDate.Year() > today.Year() || 
                         (cellDate.Year() == today.Year() && cellDate.Month() >= today.Month())
if cellDate.Before(today) && !isCurrentOrFutureMonth {
    log.Printf("Skipping past date: %s (before today: %s and not in current or future month)", 
              cellDate.Format("2006-01-02"), today.Format("2006-01-02"))
    continue
}
```

## 2. 行番号計算の不一致

`getCellPosition()`関数と`WriteSchedule()`関数で行番号の計算方法が異なります：

- `getCellPosition()`：`w.headerRow + day`（日の値を使用）
- `WriteSchedule()`：`w.headerRow + i + 1`（配列インデックスを使用）

この不一致により、取得した日付と書き込む行が一致しない可能性があります。

### 問題のあるコード
```go
// getCellPosition関数での計算
day := date.Day()
row = w.headerRow + day
```

```go
// WriteSchedule関数での計算
rowNum := w.headerRow + i + 1
```

### 修正案
両方の関数で同じ計算方法を使用するか、それぞれの目的に応じた適切な計算方法を明確に文書化する必要があります。
