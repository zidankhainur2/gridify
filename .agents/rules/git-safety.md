---
description: Aturan keselamatan Git untuk mencegah perubahan destruktif atau modifikasi remote GitHub
trigger: always_on
---

# Git Safety Rules (Keamanan Repositori Git)

Agent **DILARANG KERAS** menjalankan perintah `git` yang berpotensi merusak repositori, mengubah riwayat (history), mengubah staging/working tree secara destruktif, atau berinteraksi dengan remote repository (GitHub).

## ⛔ Perintah yang Dilarang (Forbidden Commands):
- `git push` (dilarang dalam bentuk apa pun)
- `git add`
- `git commit`
- `git rm`
- `git reset` (termasuk `--hard`, `--soft`, `--mixed`)
- `git clean`
- `git restore`
- `git revert`
- `git merge`, `git rebase`, `git cherry-pick`
- `git branch -d` / `git branch -D`
- `git stash drop` / `git stash clear`

## ✅ Ketentuan Operasi:
- Agent hanya melakukan modifikasi file lokal di workspace sesuai permintaan user.
- Staging (`git add`), committing (`git commit`), dan pushing (`git push`) ke GitHub dikelola sepenuhnya secara manual oleh USER.
- Perintah read-only (`git status`, `git diff`, `git log`) hanya digunakan untuk inspeksi bila diperlukan.
