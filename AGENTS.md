# Agent Guidelines & Repository Safety Rules

## 1. Git Safety & Repository Protection Rules (Aturan Keamanan Git)

Agent **DILARANG KERAS** menjalankan perintah `git` yang berpotensi merusak, mengubah riwayat (history), mengubah staging/working tree secara destruktif, atau melakukan perubahan pada remote repository di GitHub.

### ⛔ DILARANG KERAS (Strictly Forbidden Commands):
- **Remote Operations**:
  - `git push` (dalam bentuk apapun: `--force`, `--all`, `--tags`, atau upstream lainnya)
  - `git pull` / `git fetch` (kecuali user meminta secara eksplisit)
  - `git remote` (menambah, mengubah, atau menghapus remote)
- **Staging & Commit Operations**:
  - `git add`
  - `git commit`
  - `git rm`
- **Branch & Merge Operations**:
  - `git checkout -b`, `git switch -c`
  - `git branch` (terutama `-d`, `-D`, `-m`)
  - `git merge`, `git rebase`, `git cherry-pick`
- **Destructive Operations**:
  - `git reset` (termasuk `--soft`, `--mixed`, `--hard`)
  - `git clean` (termasuk `-f`, `-fd`, `-x`)
  - `git restore`
  - `git revert`
  - `git stash drop`, `git stash pop`, `git stash clear`

---

## 2. Working Principles & Delegation of Control

1. **Local File Editing Only**:
   - Tugas utama agent adalah membaca, menganalisis, mengedit, atau membuat file kode lokal di workspace sesuai kebutuhan user.
2. **Version Control Is Under Full User Control**:
   - Seluruh siklus Git (staging file, membuat commit pesan, mengelola branch, dan push ke GitHub) merupakan wewenang penuh **USER**.
   - Jangan pernah berinisiatif melakukan commit atau push otomatis setelah menyelesaikan tugas coding.
3. **Read-Only Inspection (Jika Dibutuhkan)**:
   - Hanya perintah read-only untuk memeriksa status yang diperbolehkan jika benar-benar relevan untuk diagnosis:
     - `git status`
     - `git diff`
     - `git log`
   - Utamakan tools sistem/filesystem bawaan (`view_file`, `list_dir`, `grep_search`) sebelum memanggil git shell command.
