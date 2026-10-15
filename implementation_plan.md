# IMPLEMENTATION PLAN — GRIDIFY UI/UX REFINEMENT, DESIGN SYSTEM OVERHAUL & FUNCTIONAL BUG FIXES

- **Nama Proyek**: Gridify — Esports & Sports Competition Bracket Operations Platform
- **Versi Dokumen**: 2.0 (Post-Revamp Re-Audit & UI/UX Refinement Plan)
- **Stack Aktual Terverifikasi**: Next.js `16.0.10` (App Router, Turbopack, React `19.2.1`, Tailwind CSS `^4.1.17`) + Go `1.24` (Modular Monolith, Chi Router, `pgx/v5`) + PostgreSQL `17`
- **Status Fase Saat Ini**: Planning & Audit Completed — **All Implementation Tasks in `TODO` Status** (Zero Source Code Mutation during Audit)

## 1. Executive Summary

### 1.1 Konteks Audit Ulang
Proses migrasi arsitektur besar (*major revamp*) sebelumnya telah berhasil memindahkan fondasi sistem Gridify dari monolit berbasis Supabase menjadi arsitektur **Next.js 16 + Go Modular Monolith + PostgreSQL 17** dengan skema relasional murni (`migrations/000001_init_schema.up.sql`), autentikasi sesi `HttpOnly` berbasis Argon2id, mesin generator bracket transaksional, dan streaming pembaruan langsung via *Server-Sent Events (SSE)*.

Namun, hasil pengujian manual pada antarmuka aplikasi serta audit mendalam baris-demi-baris terhadap seluruh kode sumber aktual (`frontend/` dan `backend/`) membuktikan bahwa **lapisan presentasi (frontend Next.js) mengalami ketidaksinkronan kontrak DTO dengan backend Go, kegagalan konfigurasi token tema Tailwind CSS v4, regresi visualisasi bracket, serta masalah kualitas desain visual**:

1. **Bug Fungsional & Console Errors Kritis (P0)**:
   - **React Unique Key Warning & Nama Tim Kosong pada `StandingsTable` (`features/bracket/components/standings-table.tsx:156`)**: Backend Go (`backend/internal/service/standings_service.go:14-25`) mengirimkan objek bersarang `participant: { id, name }`, sedangkan `StandingsTable` membaca properti tingkat atas `row.participant_id` dan `row.participant_name` yang bernilai `undefined`. Akibatnya, setiap baris `<tr>` memiliki `key={undefined}` (memicu *console error* React) dan nama peserta menjadi kosong ketika merender data `standings` dari server.
   - **Bug Visual Pemilihan Format Kompetisi (`features/tournaments/components/create-tournament-form.tsx:203-235`)**: File `frontend/app/globals.css` menggunakan `@import "tailwindcss";` (Tailwind CSS v4) tetapi **tidak mendefinisikan blok `@theme`** maupun variabel `--color-accent`, `--color-surface-*`, dan `--color-border-*`. Akibatnya, kelas utilitas `bg-accent`, `bg-accent/10`, `border-accent`, dan `text-accent` tidak menghasilkan aturan CSS sama sekali. Saat pengguna memilih format kompetisi, indikator titik radio (`bg-accent`) dan border kartu (`border-accent`) tetap transparan/tidak berubah.
   - **Kegagalan Submit Pembuatan Turnamen (`create-tournament-form.tsx:96-113` vs `tournament_handler.go:31-60`)**: Form pembuatan turnamen membungkus konfigurasi pertandingan ke dalam objek `match_rules: { best_of, home_away, has_third_place, total_matches }`, sedangkan handler Go menerapkan `dec.DisallowUnknownFields()` pada `CreateTournamentPayload` yang mengharapkan field tingkat atas (`best_of`, `home_away`, `has_third_place`, `br_total_matches`). Setiap upaya membuat turnamen melalui form UI akan ditolak dengan `400 Bad Request` (`INVALID_PAYLOAD`).
   - **Skor Pertandingan Selalu Tampil `0 - 0` di UI (`match-card.tsx`, `league-match-card.tsx`, `score-modal.tsx`)**: Backend Go mengirimkan skor pada field `score_a` dan `score_b`, namun komponen kartu pertandingan dan modal skor hanya membaca properti warisan lama `match.participant_a_score ?? match.scores?.a ?? 0`.
   - **Garis Konektor Bracket Hilang & Pengelompokan Babak Rusak (`bracket-visualizer.tsx:32-36 & 154`)**: Elemen `<path>` SVG menggunakan `stroke="var(--color-border-default)"` yang tidak pernah didefinisikan di `globals.css` sehingga garis tidak tergambar sama sekali. Selain itu, pemisahan Upper Bracket, Lower Bracket, dan Grand Final masih memakai filter angka warisan (`round_number < 0` dan `round_number === 999`), padahal backend Go menggunakan `round_number >= 1` positif dengan kolom `bracket_type` (`UPPER`, `LOWER`, `GRAND_FINAL`, `THIRD_PLACE`).
2. **Kelemahan Identitas Visual & Pola Desain Generik ("AI Slop")**:
   - Sistem warna saat ini terjebak dalam monokrom hitam-putih-abu (`#09090B`, `#121215`, `#18181B`, dan tombol primer putih pucat `bg-zinc-100 text-zinc-950`) karena seluruh kelas warna semantik (`accent`, `surface-*`, `border-*`) gagal terkompilasi oleh Tailwind v4.
   - Halaman publik dan aplikasi masih menyisakan pola dekoratif khas *template* AI: badge pengumuman tech stack (`"Arsitektur Baru: Next.js 16 + Go + Realtime SSE"` di `hero.tsx:13-21`), pill kapsul promosi berulang (`"SIAP MEMULAI KOMPETISI"`, `"SPECTATOR LIVE VIEW"`, `"PODIUM JUARA TURNAMEN"`), tautan GitHub generik (`https://github.com`) di footer, serta tautan `"Bagan Publik"` di navbar yang justru melempar pengguna ke `/login`.
   - Hero section belum menampilkan identitas visual cabang game esports yang didukung secara nyata oleh sistem.

### 1.2 Tujuan Dokumen Ini
Dokumen `IMPLEMENTATION_PLAN.md` ini menggantikan rencana sebelumnya secara menyeluruh dan berfungsi sebagai cetak biru teknis tunggal (*single source of truth*) untuk menuntaskan perbaikan bug fungsional, membangun sistem desain baru yang berkarakter (**"Tactical Broadcast Slate & Crimson Volt"**), menghapus pola *AI-slop*, membangun mesin konektor bracket SVG yang akurat, serta merombak Hero, Navbar, Footer, dan konsistensi halaman secara terukur.

## 2. Current State & Previous Revamp Assessment

### 2.1 Apa yang Sudah Berhasil pada Revamp Sebelumnya (Pertahankan)
Berdasarkan pemeriksaan kode aktual di `backend/` dan `migrations/`, fondasi backend dan database berada dalam kondisi sangat baik dan **tidak memerlukan perombakan arsitektur**:
- **Database PostgreSQL 17 Murni (`migrations/000001_init_schema.up.sql` & `000002_seed_games.up.sql`)**: Sudah ternormalisasi penuh dengan 7 tabel (`users`, `user_sessions`, `password_reset_tokens`, `games`, `tournaments`, `stages`, `participants`, `matches`, `battle_royale_match_results`), dilengkapi `FOREIGN KEY ON DELETE CASCADE`, `CHECK` constraints ketat, dan seed 10 game resmi di 4 genre (`MOBA`, `FPS`, `SPORTS`, `BATTLE_ROYALE`).
- **Mesin Algoritma & Layanan Domain Go (`backend/internal/service/**`)**:
  - Generator `single_elim.go`, `double_elim.go`, `round_robin.go`, `battle_royale.go`, dan mesin `progression.go` sudah berjalan di dalam transaksi database (`database.WithTx`), mengisi `next_match_id`, `next_match_slot`, `loser_next_match_id`, dan `loser_next_match_slot` secara eksplisit, serta memiliki unit test dan integration test yang komprehensif (`*_test.go`).
  - Layanan `standings_service.go` sudah menghitung klasemen Liga/Grup, Leaderboard Battle Royale, dan `Champions` secara terpusat di backend.
  - Kanal *realtime* `internal/realtime/hub.go` sudah memancarkan event SSE (`match_updated`, `stage_updated`, `tournament_updated`, `participant_updated`) dengan *heartbeat* 25 detik.

### 2.2 Mengapa Masalah Muncul pada Implementasi Frontend Saat Ini
Evaluasi terhadap `frontend/` mengungkapkan tiga akar penyebab mengapa hasil implementasi sebelumnya belum memenuhi standar produksi:
1. **Migrasi Setengah Jalan pada Kontrak Data Frontend (`types/api.ts` & Komponen UI)**:
   Saat backend Go selesai dibuat dengan struktur JSON baru (`score_a`/`score_b`, `game`, `playing_groups`, `br_leaderboard`, serta objek bersarang `participant` di dalam `StandingRow`), beberapa komponen di `frontend/features/` hanya diperbarui agar lulus kompilasi TypeScript dengan menambahkan properti opsional (`participant_id?: string`, `participant_a_score?: number`, `games?: Game`) di `types/api.ts`, tanpa mengubah kode komponen untuk membaca field aktual yang dikirim oleh Go!
2. **Ketidaksesuaian Konfigurasi Tailwind CSS v4 di `frontend/app/globals.css`**:
   Seluruh komponen halaman ditulis menggunakan nama token semantik (`bg-surface-base`, `bg-surface-raised`, `bg-surface-elevated`, `border-border-default`, `border-border-subtle`, `text-text-primary`, `text-text-secondary`, `text-text-muted`, `bg-accent`, `text-accent`, `border-accent`, `bg-danger`, `text-danger`), namun `app/globals.css` hanya mendefinisikan 6 variabel CSS biasa di `:root` (`--background`, `--foreground`, `--surface-raised`, `--surface-overlay`, `--surface-border`, `--surface-muted`) tanpa blok `@theme` Tailwind v4. Akibatnya, **seluruh kelas utilitas semantik tersebut diabaikan oleh compiler Tailwind v4**, meninggalkan tampilan monokrom gelap yang datar dan menghilangkan indikator interaksi (`isSelected`, `focus`, `hover`).
3. **Ketiadaan Art Direction & Kurasi Konten Produk**:
   Komponen publik (`hero.tsx`, `navbar.tsx`, `footer.tsx`, `features.tsx`, `how-it-works.tsx`) dibangun menggunakan susunan template landing page standar (badge tech stack di atas judul, grid kartu `rounded-xl` seragam dengan ikon di dalam kotak, dan link footer placeholder ke `https://github.com`) tanpa menonjolkan konten domain utama Gridify, yaitu cabang game kompetitif dan kejelasan visual bagan turnamen.

## 3. Validated Audit Findings

Sesuai instruksi audit, seluruh temuan telah divalidasi langsung terhadap kode sumber aktual dan diklasifikasikan ke dalam tiga kategori: **Confirmed Issues**, **Potential Issues**, dan **Design Improvements**.

### 3.1 Confirmed Issues (Terverifikasi 100% dari Source Code)

| ID Temuan | Komponen / Lokasi File | Kategori | Bukti Kode & Akar Masalah Aktual |
| :--- | :--- | :--- | :--- |
| **CFM-01** | `frontend/features/bracket/components/standings-table.tsx:29-42, 155-182`<br>`backend/internal/service/standings_service.go:14-25` | Functional Bug / Console Error | **React Unique Key Warning & Blank Participant Name**: Struct Go `StandingRow` mengekspor `participant: domain.Participant` (`json:"participant"`), tanpa field `participant_id` atau `participant_name`. Di `standings-table.tsx:30-42`, saat `standings` dari backend di-flatten, `r.participant_id` dan `r.participant_name` bernilai `undefined`. Baris 157 (`<tr key={row.participant_id}>`) memberikan `key={undefined}` pada setiap baris, memicu warning React `Each child in a list should have a unique "key" prop`, dan baris 177 (`{row.participant_name}`) merender teks kosong. Selain itu, baris 159 & 178 memeriksa `grpName.includes('Group')`, padahal backend Go (`round_robin.go:45`) memberi nama `"Grup A"`, `"Grup B"`, sehingga highlight & badge `Q` (Qualified) tidak pernah tampil pada mode `HYBRID_UCL`. |
| **CFM-02** | `frontend/app/globals.css:1-10`<br>`frontend/features/tournaments/components/create-tournament-form.tsx:203-235, 269-281, 335-347`<br>`frontend/features/tournaments/components/tournament-settings-form.tsx:91-117` | Functional & Visual Bug | **Competition Format, Best-Of, BR Rounds & Status Selection Invisible**: `globals.css` menggunakan `@import "tailwindcss";` (Tailwind v4) tanpa deklarasi `@theme`. Kelas `bg-accent/10`, `border-accent`, `bg-accent`, `text-accent`, `bg-surface-base`, `bg-surface-raised`, `bg-surface-elevated`, `border-border-default`, dan `border-border-subtle` tidak menghasilkan aturan CSS. Akibatnya, saat pengguna mengklik format kompetisi di `create-tournament-form.tsx:208`, state `selectedFormat` berubah di React, tetapi border kartu dan titik radio (`<div className="w-2 h-2 rounded-full bg-accent" />`) tetap transparan dan tidak terlihat sama sekali. Masalah identik terjadi pada tombol pilihan `Best Of`, putaran `Battle Royale`, dan kartu `Status Turnamen` di halaman Settings. |
| **CFM-03** | `frontend/features/tournaments/components/create-tournament-form.tsx:96-113`<br>`backend/internal/transport/http/handler/tournament_handler.go:31-60` | Critical Functional Bug | **Create Tournament Form Selalu Gagal Submit (`400 Bad Request`)**: `CreateTournamentForm` mengirim payload `{ title, game_id, format_type, description, match_rules: { best_of, home_away, has_third_place, total_matches } }`. Di backend Go, `TournamentHandler.Create` menggunakan `dec.DisallowUnknownFields()` pada `CreateTournamentPayload` yang tidak memiliki field `match_rules` (melainkan field root `best_of`, `home_away`, `has_third_place`, `br_total_matches`). Akibatnya, pembuatan turnamen dari UI selalu gagal dengan `400 INVALID_PAYLOAD`. |
| **CFM-04** | `frontend/features/bracket/components/bracket-visualizer.tsx:32-36, 58-95, 145-161` | Functional & Visual Bug | **Garis Konektor Bracket Tidak Tampil & Layout Double Elimination / Bronze Match Rusak**:<br>1. Baris 154: `<path stroke="var(--color-border-default)" />` merujuk ke variabel CSS `--color-border-default` yang tidak ada di `globals.css`, sehingga garis SVG tidak tergambar.<br>2. Baris 32–36: Memfilter `upperMatches` dengan `m.round_number > 0 && m.round_number < 999`, `lowerMatches` dengan `m.round_number < 0`, dan `grandFinalMatches` dengan `m.round_number === 999`. Padahal di backend Go (`single_elim.go`, `double_elim.go`), semua match memiliki `round_number >= 1` dan dibedakan melalui `m.bracket_type` (`UPPER`, `LOWER`, `GRAND_FINAL`, `THIRD_PLACE`). Akibatnya, pada Double Elimination, seluruh match Lower Bracket dan Grand Final tergabung ke kolom Upper Bracket Round 1..N, sedangkan seksi Lower Bracket & Grand Final selalu kosong! Pada Single Elimination dengan Bronze Match, match `THIRD_PLACE` menumpuk di kolom Final.<br>3. Baris 76–85: Titik masuk konektor (`y2`) selalu mengarah ke titik tengah vertikal kartu target (`tRect.top + tRect.height / 2`), bukan ke baris slot peserta A (atas) atau peserta B (bawah) sesuai `m.next_match_slot`, sehingga dua garis dari match berbeda menyatu di tengah sebelum menyentuh kartu dan tidak merepresentasikan slot masuk peserta secara akurat. |
| **CFM-05** | `frontend/features/bracket/components/match-card.tsx:23-24`<br>`frontend/features/bracket/components/league-match-card.tsx:23-24`<br>`frontend/features/bracket/components/score-modal.tsx:28-33`<br>`backend/internal/domain/tournament.go:77-78` | Critical Functional Bug | **Skor Pertandingan Tidak Pernah Tampil (`0 - 0`)**: Backend Go mengirimkan skor pada properti `score_a` dan `score_b`. Namun `MatchCard`, `LeagueMatchCard`, dan nilai awal input `ScoreModal` hanya membaca `match.participant_a_score ?? match.scores?.a ?? 0` tanpa membaca `match.score_a` dan `match.score_b`. Akibatnya, skor yang sudah tersimpan di database tetap tampil `0 - 0` di seluruh kartu pertandingan dan modal skor! |
| **CFM-06** | `frontend/app/dashboard/tournaments/[id]/bracket/page.tsx:48-51, 157-164`<br>`frontend/app/(public)/t/[id]/page.tsx:52-55, 235-242`<br>`frontend/features/bracket/components/br-view.tsx:21-51`<br>`frontend/features/bracket/components/br-leaderboard.tsx:31-95, 130-135`<br>`frontend/features/bracket/components/br-match-input.tsx:25, 44`<br>`frontend/features/bracket/components/br-score-modal.tsx:48, 138` | Functional Bug / Console Error | **Mode Battle Royale Rusak di Frontend**:<br>1. `bracket/page.tsx` dan `t/[id]/page.tsx` tidak meneruskan `stageViewData.br_leaderboard` maupun `stageViewData.champions` ke `<BattleRoyaleView />`.<br>2. `BRLeaderboard` membaca `s.participant_id` dan `s.participant_name` (yang bernilai `undefined` pada `BRLeaderboardRow` dari Go, berpotensi memicu React unique key warning pada `<tr key={row.teamId}>` di baris 132) atau jatuh ke fallback `m.scores?.results` yang tidak ada di Go, sehingga tabel leaderboard selalu menampilkan `0` poin untuk semua tim.<br>3. Di `battle_royale.go:53-54`, setiap game disimpan dengan `round_number = 1..N` dan `match_number = 1`. Namun `BRMatchInput` (baris 44) dan `BRScoreModal` (baris 138) menampilkan `Game #{match.match_number}`, sehingga seluruh putaran selalu bernama **`Game #1`**! Selain itu, `BRMatchInput` membaca `match.scores?.groups` alih-alih `match.playing_groups`. |
| **CFM-07** | `frontend/app/dashboard/page.tsx:176`<br>`frontend/app/dashboard/tournaments/page.tsx:176`<br>`frontend/app/dashboard/tournaments/[id]/layout.tsx:38`<br>`frontend/app/dashboard/tournaments/[id]/page.tsx:83`<br>`frontend/app/(public)/t/[id]/page.tsx:130` | Functional Bug | **Nama Game Selalu Jatuh ke Fallback `'Esports'`**: Backend Go (`domain/tournament.go:23`) mengekspor relasi game dengan key JSON tunggal `"game": { "id", "slug", "name", "genre", "platform" }`. Seluruh halaman frontend hanya membaca properti jamak `tournament.games?.name` tanpa memeriksa `tournament.game?.name`, sehingga nama game asli tidak pernah muncul di kartu maupun header turnamen. |
| **CFM-08** | `frontend/components/landing/hero.tsx:13-21`<br>`frontend/app/page.tsx:31-34`<br>`frontend/app/(public)/t/[id]/page.tsx:110-113`<br>`frontend/features/bracket/components/champions-view.tsx:75-78`<br>`frontend/components/providers/tournament/realtime-listener.tsx:14-38` | UI/UX & AI-Slop Pattern | **Elemen Dekoratif AI-Slop & Badge Promosi Internal**:<br>1. `hero.tsx:13-21` menampilkan badge `"Arsitektur Baru: Next.js 16 + Go + Realtime SSE"` yang mengekspos stack teknis internal kepada pengguna akhir.<br>2. Repetisi *floating pill badges* (`"SIAP MEMULAI KOMPETISI"`, `"SPECTATOR LIVE VIEW"`, `"PODIUM JUARA TURNAMEN"`) dan animasi masuk seragam `initial={{ opacity: 0, y: 15 }}` di hampir setiap blok.<br>3. `RealtimeListener` merender pill melayang `fixed bottom-4 right-4 z-50` (`"SSE REALTIME"`) yang menutupi konten sudut kanan bawah layar di halaman bracket dan spectator. |
| **CFM-09** | `frontend/components/landing/navbar.tsx:28-38`<br>`frontend/components/landing/footer.tsx:15-34`<br>`frontend/app/(public)/t/[id]/page.tsx:86-105` | Navigation & Product Relevance | **Masalah Struktur Navbar & Footer**:<br>1. `navbar.tsx:35-37` menampilkan menu `"Bagan Publik"` tetapi mengarah ke `href="/login"`, yang membingungkan pengguna.<br>2. `navbar.tsx` tidak memiliki menu navigasi mobile (hamburger/drawer) sama sekali (`hidden md:flex`), sehingga pengguna mobile tidak dapat mengakses menu navigasi.<br>3. `footer.tsx:26-33` menampilkan link eksternal generik ke `https://github.com` (root GitHub, bahkan bukan repositori dokumentasi pengguna) dan jargon teknis `"Modern Esports Tournament Engine"`.<br>4. Header di halaman publik `/t/[id]` selalu menampilkan tombol `"Masuk Organizer"` meskipun penyelenggara sudah login. |

### 3.2 Potential Issues (Memerlukan Validasi & Penanganan Saat Eksekusi)

| ID Potensi | Lokasi File | Deskripsi Potensi Masalah & Rencana Validasi |
| :--- | :--- | :--- |
| **POT-01** | `frontend/features/bracket/components/br-score-modal.tsx:47-75`<br>`backend/internal/transport/http/handler/bracket_handler.go`<br>`backend/internal/service/standings_service.go:84-127` | **Pre-populasi Skor Battle Royale Saat Mengedit Ulang Match yang Sudah Selesai**: Saat admin membuka kembali `BRScoreModal` untuk mengedit hasil game yang sudah `COMPLETED`, `br-score-modal.tsx:48` membaca `match.scores?.results` yang selalu kosong karena hasil BR disimpan di tabel `battle_royale_match_results`. Perlu divalidasi apakah `StageView` atau objek `Match` pada mode `BATTLE_ROYALE` perlu menyertakan `br_results` per match (atau endpoint `GET` hasil match) agar modal edit menampilkan nilai placement & kills yang sudah tersimpan sebelumnya, bukan mereset form menjadi kosong. |
| **POT-02** | `frontend/features/bracket/components/standings-table.tsx:241-243`<br>`frontend/app/dashboard/tournaments/[id]/bracket/page.tsx` | **Visibilitas Tombol `AdvanceButton` ("Lanjut ke Babak Knockout Playoff")**: Di `standings-table.tsx:118 & 241`, tombol `<AdvanceButton />` muncul setiap kali `isFinished && !isReadOnly` pada semua tabel `StandingsTable`, termasuk pada turnamen murni `ROUND_ROBIN` (Liga penuh tanpa babak Knockout) dan bahkan ketika stage Knockout pada `HYBRID_UCL` sudah pernah dibuat! Kondisi render `<AdvanceButton />` harus divalidasi agar hanya muncul jika `tournament.format_type === 'HYBRID_UCL'` dan babak Knockout belum memiliki pertandingan. |
| **POT-03** | `frontend/features/bracket/components/standings-table.tsx:131`<br>`frontend/app/dashboard/tournaments/[id]/bracket/page.tsx:136-164`<br>`frontend/app/(public)/t/[id]/page.tsx:214-243` | **Tampilan `ChampionsView` pada Format `ROUND_ROBIN` dan `BATTLE_ROYALE`**: Meskipun backend Go (`standings_service.go:477-515`) sudah menghitung objek `champions` untuk `ROUND_ROBIN` dan `BATTLE_ROYALE`, komponen `<ChampionsView />` saat ini hanya dipasang di dalam `<BracketVisualizer />` (khusus Elimination) dan tidak dirender di `<StandingsTable />` maupun `<BattleRoyaleView />` ketika turnamen selesai (`tournament.status === 'COMPLETED'`). |
| **POT-04** | `frontend/features/bracket/components/bracket-visualizer.tsx:58-108` | **Sinkronisasi Kalkulasi Ulang Konektor SVG Saat Tab / Stage Berpindah atau Font Selesai Dimuat**: `calculatePaths` saat ini hanya dijalankan pada `mount`, `setTimeout(..., 150)`, `window.resize`, dan `onScroll`. Jika pengguna beralih antar-stage (`?stage=...`), memperbarui skor via SSE, atau ukuran kontainer berubah karena pemuatan font/layout shift, koordinat DOM `getBoundingClientRect()` dapat bergeser tanpa memicu `window.resize`. Perlu digunakan `ResizeObserver` pada elemen kanvas dan kartu match. |

### 3.3 Design Improvements (Berdasarkan Arahan Produk)

| ID Desain | Area Cakupan | Arahan Perbaikan Produk & Kualitas Visual |
| :--- | :--- | :--- |
| **IMP-01** | **Color System & Theme Architecture (`app/globals.css`)** | Meninggalkan skema monokrom hitam-putih-abu yang datar dan membangun palet **"Tactical Broadcast Slate & Crimson Volt"** yang relevan dengan dunia kompetisi gaming/esports, lengkap dengan blok `@theme` Tailwind v4 untuk seluruh token permukaan, teks, border, aksen utama, aksen sekunder, serta warna status/bracket. |
| **IMP-02** | **Typography & Brand Personality (`app/layout.tsx`, `app/globals.css`)** | Mengganti kombinasi tipografi generik dengan sistem 3-font yang memiliki otoritas kompetitif: **Display Font (`Space Grotesk` / `Chakra Petch`)** untuk headline, nama turnamen, dan header babak; **Sans Font (`Plus Jakarta Sans`)** untuk keterbacaan body & form UI; serta **Tabular Mono (`JetBrains Mono`)** khusus untuk skor, ranking, seed, dan matriks statistik. |
| **IMP-03** | **Hero Section — Supported Game Logos (`components/landing/hero.tsx`, `public/games/*`)** | Merancang ulang Hero section dengan komposisi *art-directed* yang menonjolkan **10 logo game resmi yang benar-benar didukung di database** (`Mobile Legends: Bang Bang`, `Dota 2`, `League of Legends`, `Valorant`, `Counter-Strike 2`, `EA Sports FC 24`, `eFootball 2024`, `PUBG Mobile`, `Free Fire`, `Apex Legends`) beserta indikator genre dan pratinjau bagan langsung, tanpa badge promosi tech stack. |
| **IMP-04** | **Navbar & Footer Produk Kompetisi (`components/landing/navbar.tsx`, `footer.tsx`)** | Merancang ulang Navbar dengan hierarki brand, navigasi kontekstual (termasuk menu mobile responsif dan kesadaran status autentikasi), serta merombak Footer agar bersih dari link GitHub dan berfokus pada navigasi produk & daftar cabang/format kompetisi yang riil. |
| **IMP-05** | **Global De-Slop & Component Polish** | Menyederhanakan pembungkusan kartu berlapis (*card-inside-card*), menghapus animasi entrance Framer Motion yang repetitif, mengintegrasikan status koneksi SSE langsung ke header turnamen (menggantikan floating pill di kanan bawah), dan menyempurnakan tampilan tabel klasemen serta kartu pertandingan. |

## 4. Functional Bugs & Fix Specifications

### 4.1 Bug 2.A — React Unique Key Warning & Blank Names pada `StandingsTable` (serta `BRLeaderboard`)

#### A. Analisis Akar Masalah (Root Cause Analysis)
Berdasarkan investigasi pada `frontend/features/bracket/components/standings-table.tsx` dan `backend/internal/service/standings_service.go`:
1. **Apakah `row.participant_id` bernilai `null`, `undefined`, atau tidak unik?**
   **Ya — `row.participant_id` selalu bernilai `undefined` ketika prop `standings` disuplai oleh backend Go.**
   Backend Go (`standings_service.go:14-25`) mendefinisikan `StandingRow` dengan field `Participant domain.Participant` (`json:"participant"`). JSON yang dikirim oleh API `/api/v1/tournaments/:id/stages/:stageId/view` adalah:
   ```json
   {
     "group_label": "Grup A",
     "rows": [
       {
         "participant": {
           "id": "550e8400-e29b-41d4-a716-446655440000",
           "name": "ONIC Esports",
           "seed_number": 1,
           "group_name": "A"
         },
         "played": 3,
         "won": 3,
         "drawn": 0,
         "lost": 0,
         "goals_for": 6,
         "goals_against": 1,
         "goal_difference": 5,
         "points": 9,
         "rank": 1
       }
     ]
   }
   ```
   Namun, di `standings-table.tsx:29-42`, fungsi pemetaan `calculatedRows` menyalin elemen `r` mentah tanpa mengekstrak `r.participant.id` ke `participant_id` dan `r.participant.name` ke `participant_name`.
2. **Apakah terdapat participant yang muncul lebih dari sekali dalam dataset standings?**
   Tidak dalam grup yang sama; di `standings_service.go:188`, baris per grup disimpan di dalam `rowsMap := make(map[uuid.UUID]*StandingRow)` yang diindeks berdasarkan `participant.ID` unik, lalu dikelompokkan per `grpName` di `standings-table.tsx:109-114`.
3. **Mengapa React tetap memunculkan warning pada `standings-table.tsx:156`?**
   Karena di baris 155–157:
   ```tsx
   {rows.map((row, idx) => (
     <tr key={row.participant_id}>
   ```
   Setiap `row` memiliki `row.participant_id === undefined`. Ketika terdapat $\ge 2$ peserta di dalam tabel, React mendeteksi beberapa sibling `<tr>` dengan key yang sama (`undefined`) dan memunculkan Console Error `Each child in a list should have a unique "key" prop`.
4. **Apakah terdapat masalah serupa pada komponen tabel atau list lainnya?**
   **Ya, ditemukan pada `BRLeaderboard` (`features/bracket/components/br-leaderboard.tsx:33-42 & 130-133`)!**
   Jika `standings` / `brLeaderboard` diteruskan dari backend Go (`BRLeaderboardRow` juga menggunakan objek bersarang `participant: { id, name }`), `br-leaderboard.tsx:35` memetakan `teamId: s.participant_id` (`undefined`), sehingga `<tr key={row.teamId}>` di baris 132 akan mengalami React key warning yang identik.

#### B. Spesifikasi Perbaikan
1. **Penyelarasan Kontrak Backend & Normalisasi Frontend (Pertahanan Berlapis)**:
   - Di **Backend Go (`backend/internal/service/standings_service.go`)**: Tambahkan field eksplisit `ParticipantID uuid.UUID` (`json:"participant_id"`) dan `ParticipantName string` (`json:"participant_name"`) pada struct `StandingRow` dan `BRLeaderboardRow` (di samping tetap mempertahankan objek `participant` untuk kompatibilitas penuh).
   - Di **Frontend (`frontend/features/bracket/components/standings-table.tsx`)**: Pada saat menormalisasi `standings` dari backend di `calculatedRows`, ekstrak identitas peserta secara deterministik:
     ```ts
     const participantId = r.participant?.id ?? r.participant_id ?? '';
     const participantName = r.participant?.name ?? r.participant_name ?? 'Peserta';
     ```
     Gunakan `key={`${grpName}-${row.participant?.id || row.participant_id}`}` pada elemen `<tr>` (tidak menggunakan array index `idx` sebagai key utama agar identitas baris tetap stabil saat urutan peringkat berubah karena pembaruan skor SSE).
   - Perbaiki pula pengecekan highlight lolos babak grup (`standings-table.tsx:159 & 178`) dari `grpName.includes('Group')` menjadi pengecekan eksplisit apakah turnamen berformat `HYBRID_UCL` atau terdapat lebih dari 1 grup (`groupedStandings.size > 1`), sehingga 2 tim teratas di `"Grup A"`, `"Grup B"`, dst. mendapatkan indikator lolos `Q` secara konsisten.
2. **Perbaikan Serupa pada `BRLeaderboard` (`features/bracket/components/br-leaderboard.tsx`)**:
   - Terima prop `brLeaderboard?: BRLeaderboardRow[]` dari `StageViewData.br_leaderboard`.
   - Petakan `teamId: row.participant?.id || row.participant_id`, `teamName: row.participant?.name || row.participant_name`, `matchesPlayed: row.matches_played`, `wwcd: row.wins`, `placementPoints: row.placement_points`, `killPoints: row.kill_points`, `totalKills: row.total_kills`, dan `totalPoints: row.total_points`.
   - Gunakan `key={row.teamId}` yang dijamin berisi UUID unik `participant.id`.

### 4.2 Bug 2.B — Competition Format Selection Visual Selection Bug & Payload Mismatch

#### A. Analisis Akar Masalah (Root Cause Analysis)
Berdasarkan audit pada `frontend/features/tournaments/components/create-tournament-form.tsx` dan `frontend/app/globals.css`:
1. **State Management & Controlled Selection**:
   State `selectedFormat` dikelola melalui `const [selectedFormat, setSelectedFormat] = useState<TournamentFormat | ''>('')` (baris 60). Saat kartu diklik (`onClick={() => setSelectedFormat(format.id)}` pada baris 208), state React sebenarnya diperbarui dengan benar (terbukti bagian ke-3 form `"Aturan Pertandingan"` pada baris 240 `{selectedGame && selectedFormat && (...)}` muncul setelah kartu diklik).
2. **Mengapa Pilihan yang Dipilih Tidak Terlihat Secara Visual?**
   - Di `create-tournament-form.tsx:209-225`, styling kondisi aktif (`isSelected === true`) bergantung sepenuhnya pada kelas utilitas `bg-accent/10`, `border-accent`, dan `<div className="w-2 h-2 rounded-full bg-accent" />`.
   - Di `frontend/app/globals.css`, proyek menggunakan `@import "tailwindcss";` (Tailwind CSS v4), namun **tidak mendefinisikan `@theme`** untuk mendaftarkan `--color-accent`, `--color-surface-base`, `--color-surface-raised`, `--color-surface-elevated`, `--color-border-subtle`, maupun `--color-border-default`.
   - Karena token tersebut tidak terdaftar di Tailwind v4, browser tidak menerima aturan CSS apa pun untuk `bg-accent/10`, `border-accent`, maupun `bg-accent`. Baik kartu yang dipilih (`isSelected = true`) maupun yang tidak dipilih (`isSelected = false`) sama-sama memiliki background transparan, border default abu-abu tipis, dan titik radio dalam yang transparan (`background: transparent`).
3. **Masalah Aksesibilitas & Navigasi Keyboard**:
   Kartu format menggunakan elemen `<div onClick={...}>` tanpa `role="radio"`, `aria-checked`, `tabIndex`, atau handler tombol `Space`/`Enter`, sehingga tidak dapat dipilih menggunakan navigasi keyboard (`Tab`).
4. **Bug Kritis Sinkronisasi Pengiriman Form ke Backend (`create-tournament-form.tsx:96-113`)**:
   Ketika form dikirim (`handleSubmit`), nilai `bestOf`, `homeAway`, `hasThirdPlace`, dan `totalMatches` dibungkus ke dalam properti `match_rules`. Sementara itu, backend Go (`tournament_handler.go:31-60`) menerapkan `dec.DisallowUnknownFields()` pada `CreateTournamentPayload` yang hanya menerima field tingkat atas (`best_of`, `home_away`, `has_third_place`, `br_total_matches`). Akibatnya, pengiriman form gagal dengan status `400 Bad Request`.
5. **Ketidaksesuaian Daftar Genre pada `FORMAT_OPTIONS` (`create-tournament-form.tsx:20-51`)**:
   `FORMAT_OPTIONS` mencantumkan `'FIGHTING'` (genre yang tidak ada di constraint PostgreSQL `migrations/000001_init_schema.up.sql:69`) dan mengizinkan `'BATTLE_ROYALE'` pada `SINGLE_ELIMINATION` (baris 25), tetapi `handleGameChange` (baris 79–81) memaksa `setSelectedFormat('BATTLE_ROYALE')` ketika game bergenre `BATTLE_ROYALE` dipilih.

#### B. Spesifikasi Perbaikan
1. **Registrasi Penuh Token Tema di `frontend/app/globals.css` (`@theme`)**:
   Definisikan seluruh variabel warna semantik di dalam blok `@theme` Tailwind v4 sehingga `bg-accent`, `border-accent`, `text-accent`, `bg-surface-*`, dan `border-border-*` menghasilkan aturan CSS yang valid di seluruh aplikasi.
2. **Desain Ulang Komponen Pemilih Format Kompetisi (`create-tournament-form.tsx`)**:
   - Gunakan elemen `<button type="button" role="radio" aria-checked={isSelected}>` di dalam kontainer `role="radiogroup"` agar mendukung penuh fokus keyboard (`focus-visible:ring-2 focus-visible:ring-accent`) dan pembaca layar.
   - Berikan indikator visual berlapis yang sangat jelas dan proporsional saat `isSelected === true`:
     - **Border**: `border-2 border-accent` (kontras tinggi terhadap `border border-border-default` pada kartu tidak aktif).
     - **Surface Background**: `bg-accent/12` dengan subtle top-left accent bar atau glow internal yang terukur (`shadow-[inset_0_1px_0_0_rgba(255,70,85,0.25)]`).
     - **Badge & Radio Indicator**: Lingkaran indikator di kanan atas kartu yang berubah menjadi terisi penuh (`bg-accent text-white`) dengan ikon centang (`Check` 12px) serta badge tag `"DIPILIH"` (`font-mono text-[10px] uppercase tracking-wider bg-accent text-white px-2 py-0.5 rounded`).
     - **Icon & Typography**: Ikon representatif untuk tiap format (`GitBranch` untuk Single Elim, `GitFork` untuk Double Elim, `Table2` untuk Round Robin, `Trophy` untuk Hybrid UCL, `Crosshair` untuk Battle Royale) yang berubah warna dari `text-text-muted` menjadi `text-accent`, diikuti judul `text-text-primary font-bold`.
3. **Perbaikan Serupa pada Pemilih Best-Of, Putaran Battle Royale & Status Turnamen (`tournament-settings-form.tsx`)**:
   Terapkan pola tombol seleksi aktif yang sama (`border-2 border-accent bg-accent/15 text-text-primary font-bold` + indikator centang/dot aktif) pada pilihan `BO1`/`BO3`/`BO5`, jumlah ronde Battle Royale, dan kartu `Status Turnamen` (`DRAFT`, `ONGOING`, `COMPLETED`).
4. **Perbaikan Payload `handleSubmit` di `create-tournament-form.tsx` & `lib/api-client.ts`**:
   Kirimkan konfigurasi turnamen secara langsung sebagai properti tingkat atas yang sesuai dengan kontrak Go `CreateTournamentPayload`:
   ```ts
   await apiClient.tournaments.create({
     title: title.trim(),
     game_id: gameId,
     format_type: selectedFormat as TournamentFormat,
     description: description.trim() || undefined,
     best_of: parseInt(bestOf, 10) || 1,
     home_away: homeAway,
     has_third_place: hasThirdPlace,
     br_total_matches: selectedFormat === 'BATTLE_ROYALE' ? totalMatches : undefined,
   });
   ```

## 5. Updated Visual Identity & Design System

### 5.1 Design Direction: "Tactical Broadcast Slate & Crimson Volt"
Untuk menjawab masalah tampilan monokrom hitam-putih yang terasa monoton dan kurang berkarakter, identitas visual Gridify dirancang ulang dengan arah **"Tactical Broadcast Slate & Crimson Volt"** — terinspirasi dari sistem grafis siaran turnamen esports tingkat tinggi (*VCT Broadcast HUD*, *ESL Pro Tour*, *FACEIT Pro League*, dan *Challengermode*).

**Alasan Pemilihan Arah Desain**:
1. **Permukaan Slate-Carbon Berstruktur Dingin (`#0B0E14` → `#111620` → `#171E2B`)**:
   Menggantikan hitam-abu mati (`#09090B` / `#121215`) dengan *cool gunmetal slate* yang memiliki sedikit rona biru-baja gelap (*deep tactical navy-slate*). Permukaan ini memberikan kedalaman arsitektural seperti ruang kendali siaran (*broadcast control room*) dan menjaga keterbacaan tabel/bagan dalam durasi penggunaan yang panjang.
2. **Aksen Utama yang Berani & Kompetitif — Crimson Valor (`#F43F5E` / `#E11D48`)**:
   Menggantikan tombol primer putih polos (`bg-zinc-100 text-zinc-950`) yang membuat aplikasi terasa seperti template dokumentasi SaaS generik. Warna **Crimson Valor (`#F43F5E`)** memberikan energi kompetisi esports yang tegas untuk CTA utama, pilihan format yang sedang aktif, dan penanda aksi penting tanpa memerlukan gradasi neon atau efek *blur glow* murahan.
3. **Aksen Sekunder & Fungsional untuk Progresi Kompetisi**:
   - **Cyber Cyan / Electric Sky (`#0EA5E9` / `#38BDF8`)**: Digunakan khusus untuk menandai progresi jalur Upper Bracket, tahapan grup/liga aktif, dan fokus interaktif.
   - **Championship Gold (`#F59E0B` / `#FBBF24`)**: Digunakan secara eksklusif untuk Grand Final, Juara 1, dan statistik WWCD.
   - **Victory Emerald (`#10B981` / `#34D399`)**: Digunakan untuk pemenang pertandingan (`winner_id`), status `ONGOING` / `LIVE`, dan kuota lolos (`Q`).

### 5.2 Semantic Color Tokens (`frontend/app/globals.css`)

Seluruh token wajib didefinisikan di `:root` dan dipetakan ke dalam `@theme` Tailwind CSS v4 di `frontend/app/globals.css`:

| Kategori | Nama Token CSS / Kelas Tailwind v4 | Nilai Hex | Aturan Penggunaan Spesifik |
| :--- | :--- | :--- | :--- |
| **Surface (Backgrounds)** | `--color-surface-base` (`bg-surface-base`) | `#0B0E14` | Latar belakang utama halaman (`body`) dan dasar kanvas bracket. |
| **Surface (Cards)** | `--color-surface-raised` (`bg-surface-raised`) | `#111620` | Permukaan kartu utama, sidebar dashboard, navbar, dan container tabel. |
| **Surface (Interactive)** | `--color-surface-elevated` (`bg-surface-elevated`) | `#171E2B` | Permukaan input form, header tabel, modal dialog, kartu match hover, dan tab aktif. |
| **Surface (Subtle Glass)** | `--color-surface-overlay` (`bg-surface-overlay`) | `#1E2738` | Elemen badge sekunder, tooltip, dan highlight baris tabel saat hover. |
| **Borders** | `--color-border-subtle` (`border-border-subtle`) | `#1E2636` | Garis pemisah internal tabel (`divide-y`), header kartu, dan garis grid halus. |
| **Borders** | `--color-border-default` (`border-border-default`) | `#2A3447` | Border struktural kartu, input, match card standar, dan garis konektor bracket default. |
| **Borders** | `--color-border-strong` (`border-border-strong`) | `#3B4963` | Border saat hover pada kartu interaktif dan pemisah seksi utama. |
| **Typography** | `--color-text-primary` (`text-text-primary`) | `#F8FAFC` | Teks utama, judul halaman, nama tim pemenang, dan skor akhir (Kontras > 15:1). |
| **Typography** | `--color-text-secondary` (`text-text-secondary`) | `#CBD5E1` | Body copy, label form, nama peserta terjadwal, dan navigasi sekunder (Kontras > 10:1). |
| **Typography** | `--color-text-muted` (`text-text-muted`) | `#64748B` | Metadata tanggal, nomor urut match (`M#1`), teks placeholder, dan tim yang sudah gugur (Kontras ≥ 4.6:1). |
| **Primary Brand Accent** | `--color-accent` (`bg-accent`, `text-accent`, `border-accent`) | `#F43F5E` | CTA utama ("Buat Turnamen"), indikator pilihan aktif (Format, BO, Status), dan garis indikator tab aktif. |
| **Primary Accent Hover** | `--color-accent-hover` (`hover:bg-accent-hover`) | `#E11D48` | State hover/active untuk tombol primer dan link aksen utama. |
| **Primary Accent Contrast** | `--color-accent-contrast` (`text-accent-contrast`) | `#FFFFFF` | Teks di atas permukaan `bg-accent`. |
| **Secondary Tactical Accent** | `--color-secondary` (`text-secondary`, `border-secondary`) | `#38BDF8` | Indikator Upper Bracket, badge informasi stage, dan highlight konektor aktif. |
| **Status: Victory / Live** | `--color-success` (`text-success`, `bg-success`) | `#10B981` | Pemenang match, garis konektor jalur menang, badge `ONGOING`/`LIVE`, dan indikator lolos `Q`. |
| **Status: Warning / Championship** | `--color-warning` (`text-warning`, `bg-warning`) | `#F59E0B` | Status `DRAFT`, Grand Final badge, medali Juara 1, dan peringatan penguncian peserta. |
| **Status: Danger / Lower Bracket** | `--color-danger` (`text-danger`, `bg-danger`, `border-danger`) | `#EF4444` | Aksi hapus (Danger Zone), pesan error validasi, dan indikator Lower Bracket. |

**Aturan Penggunaan Warna (Strict Guardrails)**:
1. **85% Surface Neutral, 10% Functional Status, 5% Brand Accent**: Warna `accent` (`#F43F5E`) hanya digunakan untuk tindakan utama, state terpilih (`selected`), dan penanda fokus—tidak pernah digunakan sebagai warna latar kartu besar.
2. **Zero Multi-Stop Neon Gradients**: Dilarang menggunakan gradasi pelangi (`indigo-to-purple-to-pink`). Transisi permukaan hanya menggunakan perbedaan elevasi warna solid (`surface-base` → `surface-raised` → `surface-elevated`) dan border `1px` yang tajam.
3. **Semantic Bracket Differentiation**:
   - **Upper Bracket**: Aksen `secondary` (`#38BDF8` Sky/Cyan).
   - **Lower Bracket**: Aksen `amber`/`rose` (`#FB7185` / `#F59E0B`).
   - **Grand Final & Champions**: Aksen `warning` (`#F59E0B` Gold).
   - **Winning Path & Winner Slot**: Aksen `success` (`#10B981` Emerald).

### 5.3 Typography & Brand Personality (`frontend/app/layout.tsx` & `globals.css`)

Evaluasi terhadap tipografi saat ini menunjukkan bahwa penggunaan `Inter` untuk seluruh elemen membuat headline hero, nama turnamen, dan header babak terasa datar seperti dokumen kantor biasa. Untuk membangun kepribadian produk kompetisi esports yang kuat sekaligus menjaga keterbacaan tinggi, sistem tipografi dibagi menjadi **3 peran fungsional** (ketiganya berlisensi Open Font License dan dimuat secara optimal melalui `next/font/google` dengan `display: "swap"`):

1. **Display & Competitive Headlines (`--font-display`: `Space Grotesk`)**:
   - **Karakter**: Geometris, teknis, tegas, dengan proporsi aperture yang khas untuk judul kompetisi, nama turnamen, header babak bracket (`UPPER BRACKET`, `GRAND FINAL`), dan nama tim pada podium juara.
   - **Weights**: `600` (SemiBold), `700` (Bold).
2. **UI & Body Copy (`--font-sans`: `Plus Jakarta Sans`)**:
   - **Karakter**: Sangat nyaman dibaca untuk teks Bahasa Indonesia pada ukuran kecil (`12px`–`16px`), label form, deskripsi aturan, navigasi, dan nama peserta di dalam tabel/kartu match.
   - **Weights**: `400` (Regular), `500` (Medium), `600` (SemiBold), `700` (Bold).
3. **Competitive Numerals, Scores, Seeds & Timers (`--font-mono`: `JetBrains Mono`)**:
   - **Karakter**: Monospace teknis dengan `font-variant-numeric: tabular-nums` wajib aktif untuk skor pertandingan (`2 - 1`), nomor seed (`#01`), nomor match (`M#04`), statistik klasemen (`MP`, `W`, `D`, `L`, `GD`, `PTS`, `WWCD`, `KILLS`), dan tanggal.
   - **Weights**: `500` (Medium), `700` (Bold).

#### Hierarki & Skala Tipografi Responsif
| Level / Elemen | Font Family | Ukuran & Responsif | Weight & Line-Height | Letter Spacing |
| :--- | :--- | :--- | :--- | :--- |
| **Hero Display Title** | `var(--font-display)` | `text-3xl sm:text-5xl lg:text-[56px]` | `font-bold leading-[1.08]` | `tracking-[-0.03em]` |
| **Page Title (`H1`)** | `var(--font-display)` | `text-2xl md:text-3xl` | `font-bold leading-tight` | `tracking-[-0.02em]` |
| **Section / Modal Header (`H2`)** | `var(--font-display)` | `text-lg md:text-xl` | `font-semibold leading-snug` | `tracking-[-0.015em]` |
| **Card / Group Title (`H3`)** | `var(--font-sans)` | `text-sm md:text-base` | `font-semibold leading-normal` | `tracking-normal` |
| **Body Copy** | `var(--font-sans)` | `text-sm` (`14px`) | `font-normal leading-relaxed` (`1.6`) | `tracking-normal` |
| **Compact UI / Table Cell** | `var(--font-sans)` | `text-xs` (`12px`) | `font-medium leading-normal` | `tracking-normal` |
| **Numeric Score / Points** | `var(--font-mono)` | `text-sm` s/d `text-2xl` | `font-bold tabular-nums` | `tracking-tight` |
| **Technical Micro-Label / Seed** | `var(--font-mono)` | `text-[11px]` | `font-semibold uppercase` | `tracking-[0.06em]` |

## 6. Global AI-Slop Removal Strategy

Seluruh halaman telah diaudit untuk memisahkan **dekorasi AI-slop yang harus dihapus/dirombak** dari **indikator fungsional yang wajib dipertahankan**.

### 6.1 Elemen AI-Slop yang Wajib Dihapus atau Didesain Ulang

| Lokasi File & Baris | Pola AI-Slop Saat Ini | Tindakan Perbaikan |
| :--- | :--- | :--- |
| `frontend/components/landing/hero.tsx:13-21` | Badge kapsul dekoratif `<Activity /> Arsitektur Baru: Next.js 16 + Go + Realtime SSE` di atas judul utama. | **Hapus total.** Pengguna akhir (penyelenggara turnamen & pemain) tidak peduli pada framework internal. Ganti konteks bagian atas hero dengan penanda domain fungsional atau langsung masuk ke komposisi tipografi & showcase cabang game. |
| `frontend/app/page.tsx:31-34` | Pill kapsul `<Trophy /> SIAP MEMULAI KOMPETISI` di atas CTA bawah disertai heading `font-black` generik. | **Hapus pill dekoratif** dan ubah seksi CTA bawah menjadi *command bar* horizontal yang ringkas dan terstruktur (judul tegas di kiri, metrik dukungan 5 format & 10 game, serta tombol aksi di kanan). |
| `frontend/components/landing/features.tsx:13-85` | Grid 6 kartu identik (`rounded-xl` + kotak ikon `w-10 h-10` + judul + deskripsi) yang menyebutkan istilah internal seperti `"Go engine"`, `"transaksi PostgreSQL 17"`, dan `"owner_id"`, ditambah animasi `whileInView` berulang. | **Desain ulang menjadi Architectural Bento Showcase (3–4 Panel Fungsional)** yang berorientasi pada manfaat penyelenggara turnamen (1. *Multi-Format Bracket Engine*, 2. *Automated League & Battle Royale Scoring*, 3. *Instant Public Spectator & QR Broadcast*) tanpa menyebut `PostgreSQL 17` atau `owner_id`. |
| `frontend/components/landing/how-it-works.tsx:36-55` | 3 kartu langkah simetris dengan kotak angka `01`, `02`, `03` di tengah (`items-center text-center`) dan animasi Framer Motion berulang. | **Ubah menjadi Horizontal Operational Timeline / Workflow Strip** dengan orientasi rata kiri (*left-aligned*), garis penghubung langkah yang tegas, dan contoh output nyata di setiap tahap. |
| `frontend/app/(public)/t/[id]/page.tsx:108-154` | Banner penonton tengah dengan pill `<Trophy /> SPECTATOR LIVE VIEW` dan 4 kapsul info terpisah yang memakan ruang vertikal terlalu besar sebelum penonton bisa melihat bagan. | **Desain ulang menjadi Compact Tournament Broadcast Header**: satukan judul turnamen, badge nama game, status kompetisi, jumlah peserta, dan indikator koneksi `LIVE` ke dalam satu header bar yang padat sehingga bagan pertandingan langsung terlihat di paruh atas layar (*above the fold*). |
| `frontend/features/bracket/components/champions-view.tsx:75-78` | Pill `<Trophy /> PODIUM JUARA TURNAMEN` dan blok bar vertikal kosong bertuliskan angka `2`, `1`, `3` besar yang memakan tinggi 320px di atas bagan. | **Desain ulang menjadi Compact Championship Banner** horizontal 3-kolom (Gold 1st di tengah/utama, Silver 2nd, Bronze 3rd) yang elegan, padat, dan tidak mendorong bagan pertandingan terlalu jauh ke bawah. |
| `frontend/components/providers/tournament/realtime-listener.tsx:14-38` | Pill mengambang `fixed bottom-4 right-4 z-50` (`"SSE REALTIME"` / `"CONNECTING SSE..."`) yang menutupi konten kanan bawah layar. | **Ubah dari Floating Overlay menjadi Inline Status Indicator** (atau kembalikan status koneksi dari hook `useTournamentSSE` untuk ditampilkan secara elegan di bagian header halaman bracket & halaman publik sebagai badge `LIVE` / `MENYAMBUNGKAN...`). |
| `frontend/features/bracket/components/br-group-list.tsx:11-50` | Masih menggunakan kelas warna lama (`bg-indigo-500/10`, `text-indigo-400`, `bg-slate-900/50`, `border-slate-800`) dan teks bahasa Inggris (`Group Distribution`, `Teams`). | **Selaraskan ke sistem token baru** (`bg-surface-raised`, `border-border-default`, `text-secondary`) dan Bahasa Indonesia konsisten (`Distribusi Grup Lobby`, `{n} Tim`). |

### 6.2 Elemen Fungsional yang Wajib Dipertahankan & Diperkuat
- **Badge Status Turnamen (`DRAFT`, `ONGOING`, `COMPLETED`)** dan **Status Match (`SCHEDULED`, `ONGOING`, `COMPLETED`/`FINAL`)**: Dipertahankan karena memberikan informasi operasional penting.
- **Indikator Lolos (`Q`) dan Peringkat (`#1`, `#2`, `#3`)** pada `StandingsTable` dan `BRLeaderboard`: Dipertahankan dan diperbaiki logikanya.
- **Nomor Seed (`#1`..`#N`) dan Nomor Match (`M#1`)**: Dipertahankan dan diperkuat di dalam `MatchCard`.

## 7. Bracket Connector Implementation Plan

### 7.1 Audit Struktur Data Pertandingan & Progresi Aktual
Berdasarkan pemeriksaan `migrations/000001_init_schema.up.sql:167-208`, `backend/internal/service/bracket/single_elim.go`, dan `backend/internal/service/bracket/double_elim.go`, struktur data pertandingan yang dikirim oleh backend Go ke frontend memiliki karakteristik pasti berikut:

1. **Identifikasi Seksi Bracket (`bracket_type`) & Nomor Babak (`round_number`)**:
   - **Single Elimination (`single_elim.go`)**:
     - Pohon utama memiliki `bracket_type = 'UPPER'` dan `round_number = 1 .. totalRounds` (di mana `totalRounds = log2(bracketSize)`).
     - Jika `has_third_place = true` dan `totalRounds >= 2`, terdapat 1 match tambahan dengan `bracket_type = 'THIRD_PLACE'`, `round_number = totalRounds`, `match_number = 1`.
   - **Double Elimination (`double_elim.go`)**:
     - **Upper Bracket**: `bracket_type = 'UPPER'`, `round_number = 1 .. totalWBRounds`. Jumlah match pada babak $r$ adalah $\text{bracketSize} / 2^r$.
     - **Lower Bracket**: `bracket_type = 'LOWER'`, `round_number = 1 .. totalLBRounds` (di mana $\text{totalLBRounds} = 2 \times (\text{totalWBRounds} - 1)$). Jumlah match pada babak $r$ mengikuti fungsi `lbMatchCount(bracketSize, r)`:
       - Babak ganjil $r = 1, 3, 5\dots$ (*Consolidation Rounds*): Mempertemukan sesama pemenang Lower Bracket (atau loser WB R1 pada LB R1).
       - Babak genap $r = 2, 4, 6\dots$ (*Drop-Down Rounds*): Mempertemukan pemenang LB babak $(r-1)$ di **Slot A** melawan tim yang kalah (*drop-down*) dari Upper Bracket babak $(r/2 + 1)$ di **Slot B**. Jumlah match pada babak genap $r$ **sama persis** dengan babak ganjil $(r-1)$!
     - **Grand Final**: `bracket_type = 'GRAND_FINAL'`, `round_number = 1`, `match_number = 1`. Menerima pemenang Final Upper Bracket di **Slot A** (`next_match_slot = 'A'`) dan pemenang Final Lower Bracket di **Slot B** (`next_match_slot = 'B'`).
2. **Relasi Konektor Eksplisit pada Setiap Match**:
   - Setiap objek `Match` dari Go API memiliki:
     - `next_match_id?: string` (UUID pertandingan tujuan bagi pemenang).
     - `next_match_slot?: 'A' | 'B'` (Slot tujuan pada kartu target: `'A'` = baris peserta atas, `'B'` = baris peserta bawah).
     - `loser_next_match_id?: string` dan `loser_next_match_slot?: 'A' | 'B'` (Tujuan bagi tim yang kalah di Upper Bracket atau Semifinal Single Elimination).
     - `is_bye: boolean` (Menandakan pertandingan BYE di Babak 1 yang otomatis selesai tanpa dimainkan).

### 7.2 Mengapa Implementasi `BracketVisualizer` Saat Ini Gagal
1. **Kesalahan Filter `bracket_type`**: `bracket-visualizer.tsx:32-36` memfilter berdasarkan `m.round_number < 0` dan `m.round_number === 999`, sehingga seluruh match `LOWER`, `GRAND_FINAL`, dan `THIRD_PLACE` masuk ke `upperMatches` dan merusak struktur kolom maupun garis konektor.
2. **Variabel CSS Stroke Tidak Terdefinisi**: `stroke="var(--color-border-default)"` tidak dikenali sehingga `<path>` SVG tidak terlihat.
3. **Titik Masuk Konektor Tidak Membedakan Slot A dan Slot B**: Semua garis masuk ke titik tengah vertikal kartu target (`tRect.top + tRect.height / 2`), sehingga garis dari Match 1 (Slot A) dan Match 2 (Slot B) bertumpuk di satu titik yang sama dan tidak menunjuk ke baris peserta masing-masing.
4. **Tidak Ada Visualisasi Asal Drop-Down (Upper → Lower / Bronze)**: Pada Double Elimination (di mana LB babak genap memiliki jumlah match yang sama dengan LB babak ganjil di sebelah kirinya, dan Slot B diisi oleh tim yang turun dari Upper Bracket) serta Bronze Match, pengguna tidak melihat indikator dari mana peserta Slot B berasal.

### 7.3 Spesifikasi Arsitektur Konektor Bracket Baru (`features/bracket/components/bracket-visualizer.tsx` & `match-card.tsx`)

```
[Match R1 #1] (Exit Right Center) ──┐
                                    ├──► (Entry Slot A Row) [Match R2 #1]
[Match R1 #2] (Exit Right Center) ──┘    (Entry Slot B Row)
```

1. **Pemisahan Seksi Berdasarkan `bracket_type` yang Akurat**:
   - `upperMatches = matches.filter((m) => m.bracket_type === 'UPPER')`
   - `lowerMatches = matches.filter((m) => m.bracket_type === 'LOWER')`
   - `grandFinalMatches = matches.filter((m) => m.bracket_type === 'GRAND_FINAL')`
   - `thirdPlaceMatches = matches.filter((m) => m.bracket_type === 'THIRD_PLACE')`
   - Kelompokkan `upperMatches` dan `lowerMatches` berdasarkan `m.round_number` secara berurutan (`1, 2, ...`) dan urutkan setiap babak berdasarkan `m.match_number` (`1, 2, ...`).
   - Tempatkan `thirdPlaceMatches` (Perebutan Juara 3) pada panel khusus di bawah kartu Final Upper Bracket pada kolom babak terakhir Single Elimination, lengkap dengan label badge `"BRONZE MATCH • JUARA 3"`.
2. **Anchor DOM Presisi untuk Slot A dan Slot B pada `MatchCard`**:
   - Tambahkan atribut data/ID pada baris peserta di dalam `MatchCard`:
     - Elemen kartu utama: `id={`match-card-${match.id}`}`
     - Baris Peserta A: `id={`match-slot-a-${match.id}`}`
     - Baris Peserta B: `id={`match-slot-b-${match.id}`}`
   - Dengan anchor ini, fungsi kalkulasi konektor SVG menghitung:
     - **Titik Keluar (`x1, y1`)**: Sisi kanan tengah dari `match-card-${m.id}` (atau titik tengah antara baris peserta A dan B).
     - **Titik Masuk (`x2, y2`)**: Sisi kiri tengah dari `match-slot-a-${m.next_match_id}` jika `m.next_match_slot === 'A'`, atau `match-slot-b-${m.next_match_id}` jika `m.next_match_slot === 'B'` (dengan fallback ke titik tengah kartu jika `next_match_slot` kosong).
     - Hasilnya: Garis dari pertandingan ganjil (`M#1`) masuk tepat ke baris **Peserta A (atas)** pada kartu babak berikutnya, dan garis dari pertandingan genap (`M#2`) masuk tepat ke baris **Peserta B (bawah)** tanpa pernah bertabrakan dengan teks nama tim atau skor!
3. **Geometri Path Ortogonal Ber-Radius Halus (`rounded orthogonal step path`)**:
   - Buat helper murni untuk menghasilkan path SVG ortogonal dengan sudut melengkung (`r = 6px`):
     `M x1 y1 H (midX - r) Q midX y1, midX (y1 ± r) V (y2 ∓ r) Q midX y2, (midX + r) y2 H x2` (atau garis lurus `M x1 y1 H x2` jika $|y_2 - y_1| < 2\text{px}$, seperti pada transisi LB babak ganjil ke LB babak genap).
4. **Pewarnaan Semantik & Hierarki Garis Konektor**:
   - **Default Connector (`status !== 'COMPLETED'`)**: `stroke="#2A3447"` (`var(--color-border-default)`), `strokeWidth="2"`, `strokeDasharray` solid (atau `4 4` jika sumber adalah match yang belum memiliki peserta).
   - **Completed / Active Progression Connector (`m.status === 'COMPLETED' && m.winner_id && !m.is_bye`)**: `stroke="#10B981"` (`var(--color-success)` / Emerald) atau `#38BDF8` (`var(--color-secondary)`) dengan `strokeWidth="2.5"` dan `opacity="0.85"`, serta titik node kecil (`<circle r="3" />`) pada titik masuk slot target sehingga alur perjalanan tim pemenang terlihat jelas dari Babak 1 hingga Final.
   - **BYE Match Handling**: Untuk match dengan `m.is_bye === true`, tampilkan badge status `"BYE • AUTO ADVANCE"` pada `MatchCard` dengan opasitas terukur (`opacity-75`) dan garis konektor putus-putus halus (`strokeDasharray="4 4"`) menuju slot Round 2 yang sudah terisi otomatis.
   - **Drop-Down Badge Indicator (Upper → Lower & Semifinal → Bronze)**: Karena menggambar garis panjang melintasi seluruh kanvas dari Upper Bracket ke Lower Bracket akan memotong kartu pertandingan lain, tampilkan badge micro-tag asal drop-down pada slot penerima di Lower Bracket / Bronze Match (misal: `"← Loser UB R1 #1"`) cuando peserta belum ditentukan (`TBD`), sehingga hubungan antar-bracket tetap 100% jelas tanpa kekacauan garis silang.
5. **Reaktivitas Layout (`ResizeObserver`) & Responsivitas Mobile**:
   - Gunakan `ResizeObserver` pada `containerRef` dan panggil `requestAnimationFrame(calculatePaths)` saat ukuran kontainer, daftar `matches`, atau stage aktif berubah.
   - Sediakan kontrol zoom/scroll horizontal yang ergonomis pada layar kecil (`overflow-x-auto custom-scrollbar` dengan indikator geser horizontal pada mobile).

## 8. Navbar, Footer & Hero Redesign

### 8.1 Navbar Redesign (`frontend/components/landing/navbar.tsx` & Public Header)
**Hasil Audit `navbar.tsx`**:
- Menu desktop saat ini memiliki 3 link: `"Fitur"` (`#features`), `"Cara Kerja"` (`#how-it-works`), dan `"Bagan Publik"` yang mengarah ke `/login` (menyesatkan karena pengguna mengira akan melihat daftar bagan publik, tetapi justru dipaksa login).
- Tidak ada navigasi mobile (`hidden md:flex` tanpa tombol hamburger/drawer).
- Menggunakan warna monokrom `bg-zinc-100 text-zinc-950` untuk logo dan CTA.

**Spesifikasi Redesign Navbar**:
1. **Hierarki Brand**:
   - Logo Gridify menggunakan ikon geometris kompetitif dengan aksen `bg-accent text-white` (`#F43F5E`), teks `Gridify` dalam `--font-display font-bold tracking-tight`, dan sub-label teknis `TOURNAMENT OS` (`font-mono text-[10px] text-text-muted`).
2. **Navigasi Kontekstual & Jujur**:
   - Ganti link palsu `"Bagan Publik"` (`/login`) dengan navigasi ke bagian nyata di halaman utama: **`#games` ("Game Didukung")**, **`#features` ("Format & Fitur")**, dan **`#how-it-works` ("Alur Kerja")**.
3. **Adaptasi Authentication State**:
   - **Belum Login**: Tombol teks `"Masuk"` (`/login`) dan CTA utama `"Buat Turnamen"` (`/register` atau `/dashboard/create`) dengan warna `bg-accent hover:bg-accent-hover text-white`.
   - **Sudah Login (`user !== null`)**: Tampilkan chip identitas `@username`, tombol sekunder `"Buat Turnamen"` (`/dashboard/create`), dan tombol utama `"Buka Dashboard"` (`/dashboard`).
   - Terapkan pula pengecekan sesi opsional pada header halaman publik `/t/[id]` (`frontend/app/(public)/t/[id]/page.tsx`) sehingga jika penyelenggara yang sedang login membuka halaman publik, tombol kanan atas menampilkan `"Kembali ke Dashboard"` (`/dashboard/tournaments/${id}`) alih-alih `"Masuk Organizer"`.
4. **Responsivitas Mobile & Aksesibilitas**:
   - Tambahkan tombol toggle menu mobile (`md:hidden`) dengan panel dropdown/drawer beranimasi ringan untuk layar smartphone, lengkap dengan dukungan tombol `Escape` dan `focus-visible` ring.

### 8.2 Footer Redesign (`frontend/components/landing/footer.tsx`)
**Hasil Audit `footer.tsx`**:
- Menampilkan link eksternal ke `https://github.com` (baris 26–33) yang tidak relevan bagi pengguna produk turnamen esports dan bahkan hanya mengarah ke beranda GitHub.
- Terlalu tipis namun tidak informatif (hanya mengulang link `Masuk`, `Daftar`, `GitHub`).

**Spesifikasi Redesign Footer**:
1. **Hapus Total Link `GitHub` (`https://github.com`)** dan jargon developer internal.
2. **Struktur Footer 3-Kolom yang Proporsional & 100% Faktual (Tanpa Mengarang URL Palsu)**:
   - **Kolom 1 — Identitas Produk**: Brand Gridify, deskripsi singkat satu kalimat mengenai platform manajemen bagan dan klasemen kompetisi esports.
   - **Kolom 2 — Format Kompetisi Didukung (Informasi Produk Faktual)**: Daftar 5 format resmi yang didukung sistem (`Single Elimination`, `Double Elimination`, `Round Robin League`, `Hybrid Group + Knockout`, `Battle Royale Leaderboard`).
   - **Kolom 3 — Akses Cepat**: Link internal yang benar-benar ada di aplikasi (`#games` Game Didukung, `#features` Fitur Utama, `#how-it-works` Cara Kerja, `/login` Masuk Organizer, `/register` Buat Akun Organizer, `/dashboard` Dashboard).
3. **Bottom Copyright Bar**: `© 2026 Gridify — Sistem Manajemen Kompetisi Esports.`

### 8.3 Hero Section Redesign — Supported Game Logos (`frontend/components/landing/hero.tsx`)

#### A. Audit Game yang Benar-Benar Didukung Sistem
Berdasarkan verifikasi terhadap `migrations/000002_seed_games.up.sql` dan `backend/cmd/seed/main.go`, database Gridify mendukung **tepat 10 game resmi** yang terbagi ke dalam **4 genre kompetisi**:

| No | Slug (`games.slug`) | Nama Resmi (`games.name`) | Genre (`games.genre`) | Platform (`games.platform`) | Format Kompetisi yang Kompatibel |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | `mobile-legends` | **Mobile Legends: Bang Bang** | `MOBA` | Mobile | Single Elim, Double Elim, Round Robin, Hybrid UCL |
| 2 | `dota-2` | **Dota 2** | `MOBA` | PC | Single Elim, Double Elim, Round Robin, Hybrid UCL |
| 3 | `league-of-legends` | **League of Legends** | `MOBA` | PC | Single Elim, Double Elim, Round Robin, Hybrid UCL |
| 4 | `valorant` | **Valorant** | `FPS` | PC | Single Elim, Double Elim, Round Robin, Hybrid UCL |
| 5 | `counter-strike-2` | **Counter-Strike 2** | `FPS` | PC | Single Elim, Double Elim, Round Robin, Hybrid UCL |
| 6 | `ea-sports-fc` | **EA Sports FC 24** | `SPORTS` | Multiplatform | Single Elim, Double Elim, Round Robin, Hybrid UCL |
| 7 | `efootball` | **eFootball 2024** | `SPORTS` | Multiplatform | Single Elim, Double Elim, Round Robin, Hybrid UCL |
| 8 | `pubg-mobile` | **PUBG Mobile** | `BATTLE_ROYALE` | Mobile | Battle Royale Leaderboard (Multi-Lobby / Single Lobby) |
| 9 | `free-fire` | **Free Fire** | `BATTLE_ROYALE` | Mobile | Battle Royale Leaderboard (Multi-Lobby / Single Lobby) |
| 10 | `apex-legends` | **Apex Legends** | `BATTLE_ROYALE` | PC/Console | Battle Royale Leaderboard (Multi-Lobby / Single Lobby) |

#### B. Audit Aset Saat Ini & Strategi Aset Logo Game
- **Kondisi Saat Ini**: Direktori `frontend/public/` hanya berisi 5 ikon bawaan Next.js (`file.svg`, `globe.svg`, `next.svg`, `vercel.svg`, `window.svg`) yang bahkan tidak dipakai. Belum ada aset logo game di `frontend/public/games/`.
- **Strategi Aset & Fallback**:
  1. Buat direktori `frontend/public/games/` dan sediakan **10 aset logo vektor SVG (`frontend/public/games/<slug>.svg`)** yang bersih, terukur, dan dioptimalkan dengan rasio aspek seragam untuk kesepuluh game di atas (`mobile-legends.svg`, `dota-2.svg`, `league-of-legends.svg`, `valorant.svg`, `counter-strike-2.svg`, `ea-sports-fc.svg`, `efootball.svg`, `pubg-mobile.svg`, `free-fire.svg`, `apex-legends.svg`).
  2. Buat modul metadata & komponen reusable `GameLogoBadge` (`frontend/components/ui/game-logo.tsx`) yang memetakan `game.slug` ke `/games/${slug}.svg` beserta warna aksen mikro per game dan memiliki **state fallback (`onError`) otomatis**: apabila gambar gagal dimuat atau terdapat game baru di database yang belum memiliki file SVG, komponen akan merender monogram/emblem vektor berdasarkan genre (`MOBA` / `FPS` / `SPORTS` / `BATTLE_ROYALE`) sehingga tampilan tidak pernah pecah (*broken image icon*).
  3. Gunakan komponen `GameLogoBadge` ini tidak hanya di Hero Section, tetapi juga pada **pemilihan game di `CreateTournamentForm`**, **kartu turnamen di Dashboard**, dan **header halaman turnamen/spectator** agar integrasi visual game terasa menyeluruh di seluruh produk!

#### C. Visual Direction & Komposisi Art-Directed Hero Section
Alih-alih menggunakan tata letak tengah generik (1 badge + 1 judul + 1 paragraf + 2 tombol), Hero Section dirancang dengan komposisi **Editorial Split + Interactive Supported Games Roster**:
1. **Bagian Atas Hero (Headline & Value Proposition yang Berkarakter)**:
   - Tanpa badge promosi tech stack (`Arsitektur Baru: Next.js 16...` dihapus).
   - Headline dalam `--font-display` (`Space Grotesk`):
     **"Kelola Bagan & Klasemen Turnamen Esports Tanpa Kompromi."**
   - Subheadline ringkas yang langsung menjelaskan nilai produk bagi penyelenggara komunitas, kampus, dan arena kompetitif.
   - CTA Utama (`bg-accent hover:bg-accent-hover text-white font-semibold`) `"Buat Turnamen Sekarang"` didampingi tombol sekunder `"Masuk ke Dashboard"`.
2. **Komposisi Utama: Supported Games Matrix (`id="games"`)**:
   - Menampilkan **10 kartu logo game resmi** dalam tata letak grid art-directed (5 kolom di desktop `lg:grid-cols-5`, 2–3 kolom di tablet/mobile) dengan filter cepat atau pengelompokan visual per kategori genre (`MOBA`, `FPS`, `SPORTS`, `BATTLE ROYALE`).
   - Setiap tile game menampilkan:
     - **Logo Vektor Resmi Game** (`h-9 w-9` / `h-10 w-10`) yang tajam di atas panggung gelap `bg-surface-raised border border-border-default`.
     - **Nama Game** & **Tag Genre + Platform** (`MOBA • Mobile`, `FPS • PC`, dll.).
     - **Interaksi Langsung**: Saat pengguna mengklik atau mengarahkan kursor ke salah satu game pada Hero (misal memilih *Valorant*, *Mobile Legends*, *EA Sports FC 24*, atau *PUBG Mobile*), komponen `BracketPreview` di bawahnya secara dinamis menyesuaikan konteks pratinjau (menampilkan contoh bagan/format yang sesuai dengan game tersebut) dan menyediakan tombol pintas `"Buat Turnamen [Nama Game] →"`!

## 9. Global Page-by-Page UI/UX Plan

Berikut adalah pemetaan lengkap seluruh halaman dan komponen aktual di repositori ke dalam 5 kategori tindakan: **(1) Dipertahankan**, **(2) Direfactor**, **(3) Didesain Ulang**, **(4) Diperbaiki karena Bug**, dan **(5) Dihapus**.

### 9.1 Tabel Pemetaan Halaman & Komponen

| Halaman / File Komponen | Kategori Tindakan | Detail Perubahan Spesifik |
| :--- | :--- | :--- |
| `frontend/app/globals.css` | **4. Diperbaiki karena Bug** & **3. Didesain Ulang** | Tambahkan blok `@theme` Tailwind v4 lengkap untuk seluruh token warna (`surface-*`, `border-*`, `text-*`, `accent`, `secondary`, `success`, `warning`, `danger`) dan variabel tipografi (`--font-display`, `--font-sans`, `--font-mono`). |
| `frontend/app/layout.tsx` | **2. Direfactor** | Muat `Space_Grotesk` (`--font-display`), `Plus_Jakarta_Sans` (`--font-sans`), dan `JetBrains_Mono` (`--font-mono`) dari `next/font/google`, serta terapkan kelas dasar `bg-surface-base text-text-primary`. |
| `frontend/components/ui/button.tsx` | **3. Didesain Ulang** | Ganti varian `primary` dari monokrom `bg-zinc-100 text-zinc-950` menjadi `bg-accent text-accent-contrast hover:bg-accent-hover shadow-xs`; selaraskan `secondary`, `outline`, `danger`, dan `ghost` dengan token permukaan baru. |
| `frontend/components/ui/input.tsx`, `select.tsx`, `card.tsx`, `badge.tsx`, `modal.tsx` | **2. Direfactor** | Ganti warna hex hardcoded (`#121215`, `#27272A`) dengan token semantik (`bg-surface-raised`, `border-border-default`, `focus-visible:border-accent`, `focus-visible:ring-accent/30`). |
| `frontend/components/ui/confirm-dialog.tsx`, `empty-state.tsx` | **1. Dipertahankan** | Sudah menggunakan struktur token semantik (`bg-surface-raised`, `text-danger`, `bg-accent`). Akan otomatis berfungsi penuh setelah `@theme` di `globals.css` diperbaiki. |
| `frontend/components/ui/game-logo.tsx` *(File Baru)* & `frontend/public/games/*.svg` *(Aset Baru)* | **3. Didesain Ulang (Baru)** | Komponen penampil logo untuk 10 game resmi berdasarkan `slug` dengan fallback emblem genre otomatis jika aset gagal dimuat. |
| `frontend/public/{file,globe,next,vercel,window}.svg` | **5. Dihapus** | Hapus 5 file SVG bawaan template create-next-app yang tidak pernah digunakan di aplikasi. |
| `frontend/components/landing/navbar.tsx` | **3. Didesain Ulang** & **4. Diperbaiki karena Bug** | Hapus link menyesatkan `"Bagan Publik"` (`/login`), tambahkan link `#games`, tambahkan menu drawer responsif untuk mobile, dan sesuaikan tampilan berdasarkan `user`. |
| `frontend/components/landing/hero.tsx` & `bracket-preview.tsx` | **3. Didesain Ulang** | Hapus badge AI-slop `"Arsitektur Baru: Next.js 16 + Go + Realtime SSE"`. Bangun komposisi Hero baru dengan showcase interaktif 10 logo game resmi yang terhubung dengan `BracketPreview`. |
| `frontend/components/landing/features.tsx` & `how-it-works.tsx` | **3. Didesain Ulang** | Hilangkan pola 6 kartu template AI dan jargon database internal (`PostgreSQL 17`, `owner_id`). Ubah menjadi layout kapabilitas produk dan alur operasional 3 langkah yang bersih. |
| `frontend/components/landing/footer.tsx` | **3. Didesain Ulang** | Hapus link `https://github.com` dan teks developer-oriented. Buat struktur footer produk dengan daftar format kompetisi dan navigasi internal yang valid. |
| `frontend/app/page.tsx` | **2. Direfactor** | Hapus pill `"SIAP MEMULAI KOMPETISI"` pada seksi CTA bawah dan rapikan menjadi bar CTA penutup yang tajam. |
| `frontend/app/login/page.tsx`, `register/page.tsx`, `forgot-password/page.tsx`, `update-password/page.tsx` | **2. Direfactor** | Selaraskan tipografi judul dengan `font-display`, perbaiki posisi tombol ikon mata (`Eye`/`EyeOff`) pada input password agar presisi di tengah vertikal input, dan pastikan fokus input memakai `border-accent`. |
| `frontend/app/dashboard/layout.tsx` | **2. Direfactor** | Perkuat indikator menu aktif di sidebar (garis aksen kiri `border-l-2 border-accent bg-surface-elevated text-text-primary`) dan rapikan header mobile. |
| `frontend/app/dashboard/page.tsx` & `frontend/app/dashboard/tournaments/page.tsx` | **4. Diperbaiki karena Bug** & **2. Direfactor** | Perbaiki pembacaan nama game menjadi `t.game?.name || t.games?.name || 'Esports'`, tampilkan `GameLogoBadge` pada kartu turnamen, dan pastikan tombol hapus turnamen terlihat jelas pada perangkat sentuh (tidak hanya `opacity-0 group-hover:opacity-100`). |
| `frontend/app/dashboard/create/page.tsx` & `frontend/features/tournaments/components/create-tournament-form.tsx` | **4. Diperbaiki karena Bug** & **3. Didesain Ulang** | 1. Perbaiki bug visual pemilihan Format Kompetisi, Best-Of, dan jumlah putaran Battle Royale.<br>2. Tambahkan pemilih Game visual berbasis kartu/logo (`GameLogoBadge`) di samping/pengganti dropdown polos.<br>3. Perbaiki bug payload `handleSubmit` (`match_rules` → field tingkat atas `best_of`, `home_away`, `has_third_place`, `br_total_matches`) agar tidak lagi gagal `400 Bad Request`. |
| `frontend/app/dashboard/tournaments/[id]/layout.tsx` & `page.tsx` | **4. Diperbaiki karena Bug** & **2. Direfactor** | Perbaiki pembacaan `tournament.game?.name`, tampilkan logo game di header turnamen, dan integrasikan indikator status koneksi SSE di header. |
| `frontend/app/dashboard/tournaments/[id]/bracket/page.tsx` & `frontend/app/(public)/t/[id]/page.tsx` | **4. Diperbaiki karena Bug** & **2. Direfactor** | Teruskan `brLeaderboard={stageViewData?.br_leaderboard}` dan `champions={stageViewData?.champions}` ke `<BattleRoyaleView />` dan `<StandingsTable />`. Pada `/t/[id]`, hapus pill `"SPECTATOR LIVE VIEW"` dan padatkan header penonton. |
| `frontend/features/bracket/components/bracket-visualizer.tsx` & `match-card.tsx` | **4. Diperbaiki karena Bug** & **3. Didesain Ulang** | 1. Filter match berdasarkan `m.bracket_type` (`UPPER`, `LOWER`, `GRAND_FINAL`, `THIRD_PLACE`).<br>2. Perbaiki pembacaan skor `match.score_a ?? match.participant_a_score ?? 0`.<br>3. Implementasikan konektor SVG ortogonal dengan anchor per slot (`slot-a` dan `slot-b`), warna semantik untuk jalur pemenang, penanganan match `is_bye`, dan `ResizeObserver`. |
| `frontend/features/bracket/components/standings-table.tsx` & `backend/internal/service/standings_service.go` | **4. Diperbaiki karena Bug** | 1. Atasi React unique key warning dengan mengekstrak `row.participant?.id ?? row.participant_id` dan menambahkan `participant_id` & `participant_name` di serializer Go.<br>2. Perbaiki highlight lolos `Q` pada grup `"Grup A"`, `"Grup B"`.<br>3. Batasi tombol `<AdvanceButton />` hanya untuk `HYBRID_UCL` saat babak Knockout belum digenerate.<br>4. Tampilkan `<ChampionsView />` saat liga selesai. |
| `frontend/features/bracket/components/league-match-card.tsx` & `score-modal.tsx` | **4. Diperbaiki karena Bug** | Perbaiki pembacaan `match.score_a` dan `match.score_b` sehingga skor pertandingan liga dan eliminasi tampil akurat di kartu dan di dalam input modal. |
| `frontend/features/bracket/components/br-view.tsx`, `br-leaderboard.tsx`, `br-match-input.tsx`, `br-score-modal.tsx`, `br-group-list.tsx` | **4. Diperbaiki karena Bug** & **2. Direfactor** | 1. Gunakan `br_leaderboard` dari backend Go dan pastikan key `<tr>` unik (`row.participant?.id`).<br>2. Perbaiki label putaran dari `Game #{match.match_number}` (selalu `#1`) menjadi `Game #{match.round_number}` serta tampilkan `match.map_name` (`Erangel`, `Miramar`, dll.) dan `match.playing_groups`.<br>3. Sertakan hasil per match agar `BRScoreModal` menampilkan data tersimpan saat diedit ulang.<br>4. Tampilkan `<ChampionsView />` pada `BattleRoyaleView`. |
| `frontend/features/bracket/components/champions-view.tsx` | **3. Didesain Ulang** | Desain ulang menjadi Compact Podium Banner horizontal tanpa pill dekoratif dan tanpa bar vertikal raksasa. |
| `frontend/features/tournaments/components/tournament-settings-form.tsx` | **4. Diperbaiki karena Bug** | Perbaiki indikator visual kartu pemilihan `Status Turnamen` (`DRAFT`, `ONGOING`, `COMPLETED`) dengan `<button type="button" role="radio">` dan state aktif yang tegas. |
| `frontend/components/providers/tournament/realtime-listener.tsx` | **2. Direfactor** | Ubah dari pill melayang `fixed bottom-4 right-4` menjadi indikator bar/badge kompak yang tidak menutupi elemen interaktif di sudut kanan bawah layar. |
| `frontend/app/dashboard/tournaments/[id]/standings/page.tsx` | **1. Dipertahankan** | Tetap dipertahankan sebagai *redirect route* bersih (`redirect('/dashboard/tournaments/${id}/bracket')`) untuk menjaga kompatibilitas URL lama. |

## 10. Architecture & Regression Considerations

Untuk memastikan seluruh perubahan UI/UX dan perbaikan bug tidak merusak logika bisnis atau kontrak API yang sudah berjalan:
1. **Backward-Compatible API Enrichment di Backend Go**:
   - Penambahan field `participant_id` dan `participant_name` pada `StandingRow` dan `BRLeaderboardRow` di `backend/internal/service/standings_service.go` bersifat **aditif (non-breaking)**: objek `participant` tetap disertakan seperti semula.
   - Pada `TournamentHandler.Create` (`backend/internal/transport/http/handler/tournament_handler.go`), selain memperbaiki `CreateTournamentForm` di frontend agar mengirim field tingkat atas (`best_of`, `home_away`, `has_third_place`, `br_total_matches`), kita juga dapat menambahkan dukungan opsional `MatchRules *struct{...} json:"match_rules,omitempty"` pada `CreateTournamentPayload` di Go sehingga baik payload tingkat atas maupun payload dengan `match_rules` diterima dengan aman tanpa memicu error `DisallowUnknownFields()`.
   - Pada `domain.Match` (`backend/internal/domain/tournament.go`) dan `GetStageView` (`standings_service.go`), untuk mode `BATTLE_ROYALE`, kita dapat melampirkan `br_results?: BattleRoyaleMatchResult[]` pada masing-masing match sehingga ketika admin membuka kembali `BRScoreModal` untuk mengedit game yang sudah selesai, nilai `placement_rank` dan `kills` sebelumnya langsung terisi di form!
2. **Zero Database Schema Mutation**:
   - Tidak diperlukan perubahan atau penghapusan tabel pada `migrations/000001_init_schema.up.sql` maupun `000002_seed_games.up.sql`. Seluruh perbaikan memanfaatkan kolom yang sudah ada di skema PostgreSQL 17 (`score_a`, `score_b`, `bracket_type`, `round_number`, `next_match_id`, `next_match_slot`, `loser_next_match_id`, `loser_next_match_slot`, `playing_groups`, `map_name`).
3. **Preservasi Strict TypeScript & Zero External Runtime Dependencies**:
   - Tidak perlu menginstal library UI eksternal baru yang berat. Seluruh komponen menggunakan React 19, Tailwind CSS v4, `@headlessui/react`, `lucide-react`, dan SVG native dengan `typescript.ignoreBuildErrors: false`.

## 11. Phased Implementation Plan

Urutan fase disusun secara ketat berdasarkan ketergantungan (*dependency order*): fondasi Design System & Token CSS (Phase 1) harus selesai terlebih dahulu sebelum perbaikan komponen interaktif & bug fungsional (Phase 2), diikuti oleh mesin visualisasi bracket (Phase 3), redesign Hero/Navbar/Footer & aset logo game (Phase 4), penyelarasan halaman global (Phase 5), dan verifikasi akhir (Phase 6).

### Phase 1 — Design System Foundation, Color Tokens & Typography (`DS-*`)
*Fondasi CSS dan tipografi yang wajib diselesaikan pertama kali karena mendasari seluruh perbaikan visual dan state aktif komponen.*

#### `DS-001` — Implementasi Tailwind CSS v4 `@theme` & Palet Warna "Tactical Broadcast Slate & Crimson Volt"
- **Status**: `DONE`
- **Prioritas**: `P0`
- **Dependencies**: None
- **Lokasi Implementasi**:
  - `frontend/app/globals.css`
  - `frontend/components/providers/app-providers.tsx`
- **Alasan & Audit Finding**:
  Menjawab **CFM-02** dan **IMP-01**. Saat ini `globals.css` tidak memiliki blok `@theme`, sehingga seluruh kelas utilitas semantik (`bg-accent`, `border-accent`, `text-accent`, `bg-surface-base`, `bg-surface-raised`, `bg-surface-elevated`, `border-border-default`, `border-border-subtle`, `text-text-primary`, `text-text-secondary`, `text-text-muted`, `bg-danger`, `text-danger`) gagal menghasilkan CSS dan menyebabkan bug pilihan format tidak terlihat serta tampilan monokrom yang datar.
- **Deskripsi Perubahan Spesifik**:
  1. Definisikan variabel CSS di `:root` dan petakan seluruhnya di dalam blok `@theme` pada `frontend/app/globals.css`:
     - `--color-surface-base: #0B0E14`
     - `--color-surface-raised: #111620`
     - `--color-surface-elevated: #171E2B`
     - `--color-surface-overlay: #1E2738`
     - `--color-border-subtle: #1E2636`
     - `--color-border-default: #2A3447`
     - `--color-border-strong: #3B4963`
     - `--color-text-primary: #F8FAFC`
     - `--color-text-secondary: #CBD5E1`
     - `--color-text-muted: #64748B`
     - `--color-accent: #F43F5E`
     - `--color-accent-hover: #E11D48`
     - `--color-accent-contrast: #FFFFFF`
     - `--color-secondary: #38BDF8`
     - `--color-success: #10B981`
     - `--color-warning: #F59E0B`
     - `--color-danger: #EF4444`
  2. Perbarui konfigurasi warna toast di `frontend/components/providers/app-providers.tsx` agar menggunakan permukaan `#111620` dan border `#2A3447`.
- **Acceptance Criteria**:
  - Seluruh kelas `bg-accent`, `border-accent`, `text-accent`, `bg-surface-*`, `border-border-*`, dan `text-text-*` terkompilasi menjadi aturan CSS yang aktif di browser.
  - Kontras teks terhadap permukaan memenuhi standar WCAG AA ($\ge 4.5:1$ untuk teks normal, $> 10:1$ untuk teks utama).
- **Testing Requirements**:
  - Jalankan `npm run build` di `frontend/` dan verifikasi output stylesheet mengandung definisi variabel `--color-accent` dan kelas utilitas terkait.

#### `DS-002` — Implementasi Sistem Tipografi 3-Tier (`Space Grotesk`, `Plus Jakarta Sans`, `JetBrains Mono`)
- **Status**: `DONE`
- **Prioritas**: `P1`
- **Dependencies**: `DS-001`
- **Lokasi Implementasi**:
  - `frontend/app/layout.tsx`
  - `frontend/app/globals.css`
- **Alasan & Audit Finding**:
  Menjawab **IMP-02**. Penggunaan font tunggal `Inter` membuat headline dan elemen kompetitif terasa monoton dan kurang berkarakter.
- **Deskripsi Perubahan Spesifik**:
  1. Di `frontend/app/layout.tsx`, impor `Space_Grotesk` (`variable: '--font-display'`, weights `['600', '700']`), `Plus_Jakarta_Sans` (`variable: '--font-sans'`, weights `['400', '500', '600', '700']`), dan `JetBrains_Mono` (`variable: '--font-mono'`, weights `['500', '700']`) dari `next/font/google`.
  2. Daftarkan `--font-display`, `--font-sans`, dan `--font-mono` ke dalam `@theme` di `frontend/app/globals.css`, serta atur aturan default `h1, h2, .font-display { font-family: var(--font-display), sans-serif; }` dan `.font-mono { font-variant-numeric: tabular-nums; }`.
- **Acceptance Criteria**:
  - Elemen `h1`, `h2`, dan `.font-display` merender `Space Grotesk`.
  - Body text dan input form merender `Plus Jakarta Sans`.
  - Seluruh skor, nomor match, seed, dan angka tabel merender `JetBrains Mono` dengan lebar digit seragam (`tabular-nums`).
- **Testing Requirements**:
  - Verifikasi pemuatan font tanpa error kompilasi di `next build` dan inspeksi computed `font-family` pada elemen heading, body, dan skor.

#### `DS-003` — Penyelarasan Komponen UI Dasar (`Button`, `Input`, `Select`, `Card`, `Badge`, `Modal`)
- **Status**: `DONE`
- **Prioritas**: `P1`
- **Dependencies**: `DS-001`, `DS-002`
- **Lokasi Implementasi**:
  - `frontend/components/ui/button.tsx`
  - `frontend/components/ui/input.tsx`
  - `frontend/components/ui/select.tsx`
  - `frontend/components/ui/card.tsx`
  - `frontend/components/ui/badge.tsx`
  - `frontend/components/ui/modal.tsx`
- **Alasan & Audit Finding**:
  Menjawab **IMP-01** dan **CFM-02**. Komponen di `components/ui/` masih menggunakan warna hex monokrom hardcoded (`bg-zinc-100 text-zinc-950`, `bg-[#121215]`, `border-[#27272A]`) dan ring fokus abu-abu.
- **Deskripsi Perubahan Spesifik**:
  1. Pada `button.tsx`: Ubah varian `primary` menjadi `bg-accent text-accent-contrast hover:bg-accent-hover active:bg-accent-hover/90 shadow-xs`, `secondary` menjadi `bg-surface-elevated text-text-primary border border-border-default hover:border-border-strong hover:bg-surface-overlay`, `outline` menjadi `border border-border-default bg-transparent text-text-secondary hover:text-text-primary hover:bg-surface-elevated`, dan `focus-visible:ring-accent`.
  2. Pada `input.tsx`, `select.tsx`, `card.tsx`, dan `modal.tsx`: Ganti `#121215` dan `#27272A` dengan `bg-surface-raised`, `border-border-default`, `text-text-primary`, `placeholder:text-text-muted`, serta `focus-visible:border-accent focus-visible:ring-1 focus-visible:ring-accent`.
- **Acceptance Criteria**:
  - Tidak ada lagi kode warna `#121215`, `#18181B`, atau `#27272A` yang di-hardcode di dalam `frontend/components/ui/*.tsx`.
  - Tombol primer menampilkan warna Crimson Valor (`#F43F5E`) dengan teks putih tajam dan state hover/focus yang jelas.
- **Testing Requirements**:
  - Pemeriksaan `tsc --noEmit` dan verifikasi visual seluruh varian tombol, input, select, badge, dan modal.

### Phase 2 — Functional Bugs & Data Contract Fixes (`BUG-*`)
*Menuntaskan seluruh bug fungsional P0/P1 dan console errors yang ditemukan saat manual testing dan audit kode.*

#### `BUG-001` — Fix React Unique Key Warning, Blank Team Names & Group Qualification Highlight pada `StandingsTable`
- **Status**: `DONE`
- **Prioritas**: `P0`
- **Dependencies**: `DS-001`
- **Lokasi Implementasi**:
  - `backend/internal/service/standings_service.go`
  - `frontend/features/bracket/components/standings-table.tsx`
  - `frontend/types/api.ts`
- **Alasan & Audit Finding**:
  Menjawab **CFM-01** dan **POT-02** / **POT-03**. Mengatasi error `Each child in a list should have a unique "key" prop` pada `standings-table.tsx:156` yang disebabkan oleh `row.participant_id` bernilai `undefined` ketika menerima data `standings` dari backend Go, sekaligus memperbaiki nama tim yang kosong, highlight lolos grup `"Grup A"`/`"Grup B"`, dan visibilitas `<AdvanceButton />`.
- **Deskripsi Perubahan Spesifik**:
  1. Di `backend/internal/service/standings_service.go`: Tambahkan field `ParticipantID uuid.UUID` (`json:"participant_id"`) dan `ParticipantName string` (`json:"participant_name"`) pada struct `StandingRow` dan isi nilainya dari `p.ID` dan `p.Name` di `CalculateRoundRobinStandings`.
  2. Di `frontend/features/bracket/components/standings-table.tsx`:
     - Normalisasi setiap baris di `calculatedRows` sehingga `participant_id: r.participant?.id || r.participant_id` dan `participant_name: r.participant?.name || r.participant_name || 'Peserta'` selalu terdefinisi.
     - Gunakan `key={`${grpName}-${row.participant?.id || row.participant_id}`}` pada elemen `<tr>` (baris 156).
     - Perbaiki kondisi highlight lolos babak grup dari `grpName.includes('Group')` menjadi `isMultiGroup` (`groupedStandings.size > 1 || grpName.startsWith('Grup') || grpName.startsWith('Group')`).
     - Tambahkan prop `tournament?: Tournament`, `champions?: Champions | ChampionResult[]`, dan `hasKnockoutMatches?: boolean` pada `StandingsTableProps`:
       - Render `<ChampionsView />` di atas tabel klasemen apabila turnamen berformat `ROUND_ROBIN` telah selesai (`champions` tersedia).
       - Hanya tampilkan `<AdvanceButton />` apabila `tournament?.format_type === 'HYBRID_UCL' && !hasKnockoutMatches && isFinished && !isReadOnly`.
- **Acceptance Criteria**:
  - 0 React key warning di browser console saat membuka halaman klasemen `ROUND_ROBIN` maupun `HYBRID_UCL`.
  - Nama seluruh peserta tampil lengkap dan akurat di setiap baris tabel klasemen.
  - Peringkat 1 dan 2 pada setiap grup di mode `HYBRID_UCL` menampilkan highlight baris dan badge `Q`.
  - Tombol `"Lanjut ke Babak Knockout Playoff"` hanya tampil pada mode `HYBRID_UCL` saat fase grup selesai dan babak Knockout belum dibuat.
- **Testing Requirements**:
  - Jalankan `go test ./internal/service/...` di backend dan uji render `StandingsTable` dengan dataset `GroupStanding[]` dari API Go.

#### `BUG-002` — Fix Visual Selection Feedback, Keyboard Accessibility & Payload Submission pada `CreateTournamentForm` dan `TournamentSettingsForm`
- **Status**: `DONE`
- **Prioritas**: `P0`
- **Dependencies**: `DS-001`, `DS-003`
- **Lokasi Implementasi**:
  - `frontend/features/tournaments/components/create-tournament-form.tsx`
  - `frontend/features/tournaments/components/tournament-settings-form.tsx`
  - `backend/internal/transport/http/handler/tournament_handler.go`
- **Alasan & Audit Finding**:
  Menjawab **CFM-02** dan **CFM-03**. Memperbaiki bug di mana pilihan format kompetisi, seri Best-Of, putaran Battle Royale, dan status turnamen tidak terlihat saat dipilih, serta memperbaiki bug `400 Bad Request` (`INVALID_PAYLOAD`) saat menyimpan turnamen baru.
- **Deskripsi Perubahan Spesifik**:
  1. Di `frontend/features/tournaments/components/create-tournament-form.tsx`:
     - Ubah kartu pemilihan format kompetisi menjadi elemen `<button type="button" role="radio" aria-checked={isSelected}>` di dalam `<div role="radiogroup">`.
     - Terapkan styling state aktif yang kontras dan jelas:
       - **Selected (`isSelected === true`)**: `bg-accent/12 border-2 border-accent shadow-sm ring-1 ring-accent/30`, ikon format berwarna `text-accent`, badge status `"AKTIF"` (`bg-accent text-white text-[10px] font-mono font-bold px-2 py-0.5 rounded`), dan lingkaran radio terisi `bg-accent border-accent text-white` dengan ikon `Check`.
       - **Unselected (`isSelected === false`)**: `bg-surface-base border border-border-default hover:border-border-strong hover:bg-surface-elevated/60`.
       - **Focus**: `focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent`.
     - Terapkan indikator state aktif yang tegas (`border-2 border-accent bg-accent/15 text-text-primary font-bold`) pada tombol pilihan `Best Of` (`BO1`, `BO3`, `BO5`) dan jumlah putaran `Battle Royale` (`3`–`8` Round).
     - Bersihkan `FORMAT_OPTIONS` dari genre fiktif `'FIGHTING'` dan sesuaikan `validGenres` dengan 4 genre resmi database (`MOBA`, `FPS`, `SPORTS`, `BATTLE_ROYALE`).
     - Perbaiki `handleSubmit` agar mengirimkan field tingkat atas (`best_of`, `home_away`, `has_third_place`, `br_total_matches`) secara langsung tanpa membungkusnya ke dalam `match_rules` yang ditolak oleh `DisallowUnknownFields()`.
  2. Di `backend/internal/transport/http/handler/tournament_handler.go`:
     - Tambahkan dukungan kompatibilitas mundur pada `CreateTournamentPayload` untuk membaca field opsional `MatchRules` jika dikirim oleh klien lama, sambil tetap memprioritaskan field tingkat atas (`best_of`, `home_away`, `has_third_place`, `br_total_matches`).
  3. Di `frontend/features/tournaments/components/tournament-settings-form.tsx`:
     - Ubah kartu pemilihan `Status Turnamen` (`DRAFT`, `ONGOING`, `COMPLETED`) menjadi `<button type="button" role="radio" aria-checked={isSelected}>` dengan border `border-2 border-accent bg-accent/12` dan indikator aktif yang jelas.
- **Acceptance Criteria**:
  - Ketika pengguna memilih format kompetisi, kartu yang dipilih langsung menampilkan border `2px` berwarna `accent` (`#F43F5E`), latar belakang ter-highlight, ikon centang aktif, dan badge indikator.
  - Pengguna dapat berpindah antar-pilihan format menggunakan tombol `Tab` dan memilih dengan `Enter`/`Space`.
  - Pengiriman form `Buat Turnamen Sekarang` berhasil (`201 Created`) untuk kelima format kompetisi dan menyimpan `best_of`, `home_away`, `has_third_place`, serta `br_total_matches` secara akurat di PostgreSQL.
- **Testing Requirements**:
  - Uji unit/integrasi pada `POST /api/v1/tournaments` di backend Go (`tournament_test.go`) dan uji alur pembuatan turnamen di frontend.

#### `BUG-003` — Fix Score Display & Input Binding (`score_a` / `score_b`) serta Game Name Resolution (`tournament.game`)
- **Status**: `DONE`
- **Prioritas**: `P0`
- **Dependencies**: `DS-001`
- **Lokasi Implementasi**:
  - `frontend/features/bracket/components/match-card.tsx`
  - `frontend/features/bracket/components/league-match-card.tsx`
  - `frontend/features/bracket/components/score-modal.tsx`
  - `frontend/features/bracket/components/standings-table.tsx`
  - `frontend/app/dashboard/page.tsx`
  - `frontend/app/dashboard/tournaments/page.tsx`
  - `frontend/app/dashboard/tournaments/[id]/layout.tsx`
  - `frontend/app/dashboard/tournaments/[id]/page.tsx`
  - `frontend/app/(public)/t/[id]/page.tsx`
- **Alasan & Audit Finding**:
  Menjawab **CFM-05** dan **CFM-07**. Saat ini seluruh kartu pertandingan dan modal skor mengabaikan `match.score_a` dan `match.score_b` dari backend Go (sehingga selalu tampil `0 - 0`), dan seluruh halaman membaca `tournament.games?.name` alih-alih `tournament.game?.name` (sehingga nama game selalu jatuh ke fallback `'Esports'`).
- **Deskripsi Perubahan Spesifik**:
  1. Pada `match-card.tsx`, `league-match-card.tsx`, `score-modal.tsx`, dan fallback `standings-table.tsx`:
     - Ubah pembacaan skor menjadi:
       ```ts
       const scoreA = match.score_a ?? match.participant_a_score ?? match.scores?.a ?? 0;
       const scoreB = match.score_b ?? match.participant_b_score ?? match.scores?.b ?? 0;
       ```
     - Pada `ScoreModal`, tambahkan `useEffect` yang menyinkronkan `scoreA` dan `scoreB` setiap kali `match` atau `isOpen` berubah agar modal selalu menampilkan skor terkini.
  2. Pada `dashboard/page.tsx`, `dashboard/tournaments/page.tsx`, `dashboard/tournaments/[id]/layout.tsx`, `dashboard/tournaments/[id]/page.tsx`, dan `(public)/t/[id]/page.tsx`:
     - Ubah resolusi nama game menjadi `const game = tournament.game || tournament.games;` dan tampilkan `game?.name || 'Esports'`.
- **Acceptance Criteria**:
  - Skor pertandingan yang tersimpan di database tampil dengan benar pada `MatchCard`, `LeagueMatchCard`, dan `ScoreModal`.
  - Nama game asli (misal `Valorant`, `Dota 2`, `Mobile Legends: Bang Bang`, `EA Sports FC 24`, `PUBG Mobile`) tampil di seluruh kartu dashboard, header turnamen, dan halaman publik.
- **Testing Requirements**:
  - Verifikasi tampilan skor pada turnamen hasil `make seed` dan setelah update skor melalui `ScoreModal`.

#### `BUG-004` — Fix Battle Royale Leaderboard, Round Numbering, Playing Groups & Match Results Pre-population
- **Status**: `DONE`
- **Prioritas**: `P0`
- **Dependencies**: `BUG-001`
- **Lokasi Implementasi**:
  - `backend/internal/domain/tournament.go`
  - `backend/internal/service/standings_service.go`
  - `frontend/types/api.ts`
  - `frontend/app/dashboard/tournaments/[id]/bracket/page.tsx`
  - `frontend/app/(public)/t/[id]/page.tsx`
  - `frontend/features/bracket/components/br-view.tsx`
  - `frontend/features/bracket/components/br-leaderboard.tsx`
  - `frontend/features/bracket/components/br-match-input.tsx`
  - `frontend/features/bracket/components/br-score-modal.tsx`
  - `frontend/features/bracket/components/br-group-list.tsx`
- **Alasan & Audit Finding**:
  Menjawab **CFM-06** dan **POT-01** / **POT-03**. Mode Battle Royale saat ini tidak meneruskan `br_leaderboard` ke komponen UI, menampilkan `Game #1` untuk semua putaran karena membaca `match.match_number` (bukan `match.round_number`), tidak menampilkan `map_name` & `playing_groups`, serta mengosongkan `BRScoreModal` saat mengedit ulang hasil pertandingan.
- **Deskripsi Perubahan Spesifik**:
  1. Di `backend/internal/domain/tournament.go` & `backend/internal/service/standings_service.go`:
     - Tambahkan `ParticipantID uuid.UUID` (`json:"participant_id"`) dan `ParticipantName string` (`json:"participant_name"`) pada `BRLeaderboardRow`.
     - Tambahkan field opsional `BRResults []BattleRoyaleMatchResult` (`json:"br_results,omitempty"`) pada `domain.Match`, dan di `GetStageView` (ketika `tournament.FormatType == "BATTLE_ROYALE"`), kelompokkan `brResults` per `MatchID` ke dalam masing-masing `view.Matches[i].BRResults`.
  2. Di `frontend/app/dashboard/tournaments/[id]/bracket/page.tsx` dan `frontend/app/(public)/t/[id]/page.tsx`:
     - Teruskan `brLeaderboard={stageViewData?.br_leaderboard}`, `champions={stageViewData?.champions}`, dan `tournament={tournament}` ke `<BattleRoyaleView />`.
  3. Di `frontend/features/bracket/components/br-view.tsx` & `br-leaderboard.tsx`:
     - Render `<ChampionsView />` di bagian atas `BattleRoyaleView` jika `champions` sudah tersedia dan seluruh putaran selesai.
     - Di `BRLeaderboard`, gunakan `brLeaderboard` dari backend Go, tampilkan kolom `#`, `Nama Tim`, `Match`, `WWCD`, `Placement Pts`, `Kill Pts` (`Kills`), dan `TOTAL PTS`, dengan `key={row.teamId}` yang dijamin berisi UUID unik `participant.id`.
  4. Di `frontend/features/bracket/components/br-match-input.tsx` & `br-score-modal.tsx`:
     - Gunakan nomor putaran `const gameNum = match.round_number || match.match_number;` sehingga kartu dan modal menampilkan **`Game #1`**, **`Game #2`**, **`Game #3`**, dst., dilengkapi badge nama peta (`match.map_name`, misal `Erangel`, `Miramar`, `Sanhok`) dan daftar grup aktif (`match.playing_groups`).
     - Di `BRScoreModal`, baca hasil yang sudah tersimpan dari `match.br_results || match.scores?.results || []` agar saat admin mengedit game yang sudah selesai, nilai `placement_rank` dan `kills` sebelumnya langsung terisi di input.
- **Acceptance Criteria**:
  - Tabel `Klasemen Global Battle Royale` menampilkan poin, WWCD, dan kills secara akurat tanpa React key warning.
  - Kartu jadwal Battle Royale menampilkan `Game #1` s/d `Game #N` beserta nama map (`Erangel`, `Miramar`, dll.) dan grup yang bermain.
  - Membuka kembali `BRScoreModal` pada game yang sudah `COMPLETED` menampilkan kembali peringkat dan jumlah kill yang telah disimpan sebelumnya.
- **Testing Requirements**:
  - Jalankan `go test ./internal/service/...` dan verifikasi tampilan turnamen demo Battle Royale (`[Demo] PUBG Mobile Pro League Finals`).

### Phase 3 — Bracket Visualization & Connector Engine (`BRACKET-*`)
*Merombak sistem tata letak bracket dan konektor SVG agar akurat untuk Single Elimination, Double Elimination, Bronze Match, dan BYE.*

#### `BRACKET-001` — Implementasi Pemisahan `bracket_type` & Layout Presisi untuk Single/Double Elimination, Grand Final & Bronze Match
- **Status**: `DONE`
- **Prioritas**: `P0`
- **Dependencies**: `DS-001`, `BUG-003`
- **Lokasi Implementasi**:
  - `frontend/features/bracket/components/bracket-visualizer.tsx`
  - `frontend/features/bracket/components/match-card.tsx`
- **Alasan & Audit Finding**:
  Menjawab **CFM-04**. Mengganti filter warisan `round_number < 0` dan `round_number === 999` dengan `m.bracket_type` (`UPPER`, `LOWER`, `GRAND_FINAL`, `THIRD_PLACE`), serta menambahkan anchor DOM spesifik per slot peserta (`A` dan `B`) pada `MatchCard`.
- **Deskripsi Perubahan Spesifik**:
  1. Di `frontend/features/bracket/components/match-card.tsx`:
     - Tambahkan `id={`match-slot-a-${match.id}`}` pada baris Peserta 1 (Slot A) dan `id={`match-slot-b-${match.id}`}` pada baris Peserta 2 (Slot B).
     - Tampilkan nomor seed peserta (`#1`..`#N` dari `p1?.seed_number` / `p2?.seed_number`) dalam kotak monospace kecil di sebelah kiri nama peserta.
     - Tangani tampilan match BYE (`match.is_bye === true`): tampilkan status `"BYE"` di header kartu dan label `"AUTO ADVANCE (BYE)"` pada slot kosong, serta nonaktifkan modal edit skor pada match BYE.
     - Tambahkan prop opsional `slotAHint?: string` dan `slotBHint?: string` untuk menampilkan petunjuk asal peserta (misal `"Pemenang M#1"`, `"Loser UB R1 #1"`, atau `"Loser Semifinal #1"`) ketika slot masih kosong (`!p1` atau `!p2`).
  2. Di `frontend/features/bracket/components/bracket-visualizer.tsx`:
     - Pisahkan daftar match secara eksplisit menggunakan `m.bracket_type`:
       - `upperMatches = matches.filter((m) => m.bracket_type === 'UPPER')`
       - `lowerMatches = matches.filter((m) => m.bracket_type === 'LOWER')`
       - `grandFinalMatches = matches.filter((m) => m.bracket_type === 'GRAND_FINAL')`
       - `thirdPlaceMatches = matches.filter((m) => m.bracket_type === 'THIRD_PLACE')`
     - Berikan judul babak yang akurat untuk Upper Bracket (`Babak 1`, `Quarterfinal`, `Semifinal`, `Final Upper Bracket` / `Grand Final`) dan Lower Bracket (`LB Babak 1`, `LB Babak 2 (Drop-Down)`, ..., `Final Lower Bracket`).
     - Tampilkan `thirdPlaceMatches` pada blok khusus di bawah kartu Final pada Single Elimination apabila `thirdPlaceMatches.length > 0`.
- **Acceptance Criteria**:
  - Pada turnamen Double Elimination, Upper Bracket, Lower Bracket, dan Grand Final dirender pada seksi masing-masing tanpa pernah bercampur di kolom yang salah.
  - Pada turnamen Single Elimination dengan `has_third_place = true`, pertandingan perebutan Juara 3 tampil terpisah dengan label `"BRONZE MATCH • PEREBUTAN JUARA 3"`.
  - Pertandingan BYE dan slot yang menunggu lawan menampilkan informasi asal progresi yang jelas.
- **Testing Requirements**:
  - Verifikasi visual dan struktural pada turnamen Single Elimination (4, 6, 8, 16 peserta + Bronze Match) dan Double Elimination (4, 6, 8 peserta).

#### `BRACKET-002` — Implementasi Mesin Konektor SVG Ortogonal Berbasis Slot (`next_match_slot`) & `ResizeObserver`
- **Status**: `DONE`
- **Prioritas**: `P0`
- **Dependencies**: `BRACKET-001`
- **Lokasi Implementasi**:
  - `frontend/features/bracket/components/bracket-visualizer.tsx`
- **Alasan & Audit Finding**:
  Menjawab **CFM-04** dan **POT-04**. Memperbaiki garis konektor SVG yang sebelumnya tidak terlihat karena `stroke="var(--color-border-default)"` tidak terdefinisi, serta memastikan setiap garis menghubungkan kartu sumber tepat ke baris slot target (`A` atau `B`) tanpa memotong teks atau kartu lain.
- **Deskripsi Perubahan Spesifik**:
  1. Di `calculatePaths` (`bracket-visualizer.tsx`):
     - Untuk setiap match `m` yang memiliki `m.next_match_id`:
       - Cari elemen sumber `document.getElementById(`match-card-${m.id}`)`.
       - Cari elemen slot target berdasarkan `m.next_match_slot`:
         - Jika `m.next_match_slot === 'A'`, targetkan `document.getElementById(`match-slot-a-${m.next_match_id}`)`.
         - Jika `m.next_match_slot === 'B'`, targetkan `document.getElementById(`match-slot-b-${m.next_match_id}`)`.
         - Fallback ke `document.getElementById(`match-card-${m.next_match_id}`)`.
       - Hitung koordinat relatif terhadap `containerRef.current`:
         - `x1 = sRect.right - cRect.left + container.scrollLeft`
         - `y1 = sRect.top + sRect.height / 2 - cRect.top + container.scrollTop`
         - `x2 = tSlotRect.left - cRect.left + container.scrollLeft`
         - `y2 = tSlotRect.top + tSlotRect.height / 2 - cRect.top + container.scrollTop`
       - Bangun path SVG ortogonal presisi dengan sudut melengkung (`radius = 6px`) pada titik belok tengah `midX = x1 + (x2 - x1) / 2`.
       - Tandai status konektor: `isCompleted = m.status === 'COMPLETED' && Boolean(m.winner_id)`, `isBye = m.is_bye`.
  2. Render `<svg>` dengan 2 layer urutan:
     - Layer 1: Garis konektor default (`stroke="#2A3447"`, `strokeWidth="2"`, atau `strokeDasharray="4 4"` untuk `isBye`).
     - Layer 2 (di atas Layer 1): Garis konektor pertandingan yang sudah selesai (`stroke="#10B981"` untuk pemenang reguler, `strokeWidth="2.5"`), ditambah titik terminal `<circle cx={x2 - 2} cy={y2} r="3" fill="#10B981" />` pada titik masuk slot target.
  3. Pasang `ResizeObserver` pada `containerRef.current` dan panggil `calculatePaths` melalui `requestAnimationFrame` agar posisi garis selalu sinkron saat ukuran layar, font, atau stage berubah.
- **Acceptance Criteria**:
  - Seluruh garis konektor terlihat jelas dengan kontras terukur di atas kanvas `#0B0E14`.
  - Garis dari pertandingan ganjil masuk ke baris Peserta A (atas) dan garis dari pertandingan genap masuk ke baris Peserta B (bawah) pada kartu babak berikutnya.
  - Jalur kemenangan tim yang sudah menyelesaikan pertandingan ter-highlight dengan warna hijau emerald (`#10B981`).
  - Tidak ada garis yang bertabrakan dengan teks nama peserta atau skor, baik pada Single Elimination, Double Elimination (termasuk transisi LB ganjil ke LB genap dan menuju Grand Final), maupun ukuran peserta non-pangkat-dua (BYE).
- **Testing Requirements**:
  - Uji pada viewport Desktop (`1440px`), Tablet (`768px`), dan Mobile (`375px`) dengan horizontal scroll aktif.

### Phase 4 — Hero Section Supported Game Logos, Navbar, Footer & Global AI-Slop Removal (`HERO-*`, `NAV-*`, `UX-*`)
*Membangun karakter visual halaman utama, mengintegrasikan aset logo 10 game resmi, serta membersihkan pola AI-slop.*

#### `HERO-001` — Pembuatan Aset Logo Vektor 10 Game Resmi & Komponen `GameLogoBadge` dengan Fallback
- **Status**: `DONE`
- **Prioritas**: `P1`
- **Dependencies**: `DS-001`
- **Lokasi Implementasi**:
  - `frontend/public/games/mobile-legends.svg` *(File Baru)*
  - `frontend/public/games/dota-2.svg` *(File Baru)*
  - `frontend/public/games/league-of-legends.svg` *(File Baru)*
  - `frontend/public/games/valorant.svg` *(File Baru)*
  - `frontend/public/games/counter-strike-2.svg` *(File Baru)*
  - `frontend/public/games/ea-sports-fc.svg` *(File Baru)*
  - `frontend/public/games/efootball.svg` *(File Baru)*
  - `frontend/public/games/pubg-mobile.svg` *(File Baru)*
  - `frontend/public/games/free-fire.svg` *(File Baru)*
  - `frontend/public/games/apex-legends.svg` *(File Baru)*
  - `frontend/components/ui/game-logo.tsx` *(File Baru)*
  - Hapus: `frontend/public/{file,globe,next,vercel,window}.svg`
- **Alasan & Audit Finding**:
  Menjawab **IMP-03** dan **Section 7 (Hero Section Redesign — Supported Game Logos)**. Menyediakan representasi visual resmi yang tajam untuk ke-10 game yang ada di `migrations/000002_seed_games.up.sql` beserta mekanisme fallback yang tangguh.
- **Deskripsi Perubahan Spesifik**:
  1. Buat 10 file SVG teroptimasi di `frontend/public/games/<slug>.svg` yang merepresentasikan mark/emblem ikonik dari masing-masing game resmi di database (`mobile-legends`, `dota-2`, `league-of-legends`, `valorant`, `counter-strike-2`, `ea-sports-fc`, `efootball`, `pubg-mobile`, `free-fire`, `apex-legends`).
  2. Bangun komponen `GameLogoBadge` dan konstanta `SUPPORTED_GAMES_META` di `frontend/components/ui/game-logo.tsx` yang menerima `slug?: string`, `name?: string`, `genre?: string`, dan `size?: 'sm' | 'md' | 'lg'`:
     - Merender `<img>` / SVG dari `/games/${slug}.svg`.
     - Jika `slug` tidak dikenali atau gambar gagal dimuat (`onError`), otomatis merender badge monogram/ikon genre (`MOBA`, `FPS`, `SPORTS`, `BATTLE_ROYALE`) sehingga tidak pernah menampilkan broken image.
  3. Hapus 5 file SVG template bawaan Next.js (`file.svg`, `globe.svg`, `next.svg`, `vercel.svg`, `window.svg`) dari `frontend/public/`.
- **Acceptance Criteria**:
  - Ke-10 game di `migrations/000002_seed_games.up.sql` memiliki file SVG yang valid di `frontend/public/games/`.
  - `GameLogoBadge` menampilkan logo yang sesuai untuk ke-10 game dan menampilkan fallback rapi jika diberikan `slug` yang tidak dikenal.
- **Testing Requirements**:
  - Verifikasi seluruh 10 file SVG dapat dimuat dengan status `200 OK` dan uji fallback `GameLogoBadge` dengan slug dummy.

#### `HERO-002` — Redesign Hero Section & Interactive Bracket Preview dengan Roster 10 Game Kompetitif
- **Status**: `DONE`
- **Prioritas**: `P1`
- **Dependencies**: `DS-001`, `DS-002`, `HERO-001`
- **Lokasi Implementasi**:
  - `frontend/components/landing/hero.tsx`
  - `frontend/components/landing/bracket-preview.tsx`
- **Alasan & Audit Finding**:
  Menjawab **CFM-08** dan **IMP-03**. Menghapus badge dekoratif `"Arsitektur Baru: Next.js 16 + Go + Realtime SSE"` dan merombak Hero section menjadi komposisi art-directed yang menampilkan ke-10 game yang didukung beserta pratinjau bracket interaktif.
- **Deskripsi Perubahan Spesifik**:
  1. Di `frontend/components/landing/hero.tsx`:
     - **Hapus total** blok `<motion.div>` badge `"Arsitektur Baru: Next.js 16 + Go + Realtime SSE"`.
     - Bangun komposisi Hero baru yang terdiri dari:
       - **Header Editorial Tajam**: Headline dalam `--font-display` (`Space Grotesk`), subheadline fokus produk, dan tombol CTA utama (`Buat Turnamen Sekarang` dengan `bg-accent` & `Masuk ke Dashboard`).
       - **Supported Games Showcase Grid (`id="games"`)**: Menampilkan ke-10 game resmi menggunakan `GameLogoBadge` dalam grid terstruktur dengan filter genre (`Semua (10)`, `MOBA (3)`, `FPS (2)`, `SPORTS (2)`, `BATTLE ROYALE (3)`).
       - Setiap kartu game pada Hero bersifat interaktif: ketika diklik, kartu tersebut menjadi aktif (`border-accent bg-accent/10`) dan memperbarui konteks pratinjau di `BracketPreview` di bawahnya (misal menampilkan simulasi bracket/leaderboard untuk game yang dipilih beserta format yang didukung).
  2. Di `frontend/components/landing/bracket-preview.tsx`:
     - Terima prop game aktif dari `Hero` (default: `Valorant` / `Mobile Legends: Bang Bang`), tampilkan logo game pada header pratinjau, gambar garis konektor SVG ortogonal nyata antara kolom Semifinals dan Grand Final pada komponen pratinjau, dan ganti warna hardcoded `#121215`/`#27272A` dengan token permukaan baru.
- **Acceptance Criteria**:
  - Tidak ada lagi badge `"Arsitektur Baru: Next.js 16 + Go + Realtime SSE"` di Hero section.
  - Hero section menampilkan ke-10 logo game resmi yang didukung aplikasi secara seimbang, proporsional, dan responsif di desktop, tablet, dan mobile.
  - Interaksi memilih game pada Hero langsung memperbarui konteks `BracketPreview`.
- **Testing Requirements**:
  - Uji responsivitas Hero section pada layar `375px`, `768px`, dan `1440px`, serta uji interaksi pemilihan game dan filter genre.

#### `NAV-001` — Redesign Navbar & Footer Produk (Hapus Link GitHub & Perbaiki Navigasi)
- **Status**: `DONE`
- **Prioritas**: `P1`
- **Dependencies**: `DS-001`, `DS-002`, `HERO-001`
- **Lokasi Implementasi**:
  - `frontend/components/landing/navbar.tsx`
  - `frontend/components/landing/footer.tsx`
- **Alasan & Audit Finding**:
  Menjawab **CFM-09** dan **IMP-04**. Memperbaiki hierarki Navbar, menambahkan navigasi mobile, menghapus link menyesatkan `"Bagan Publik"` (`/login`), dan menghapus link `GitHub` (`https://github.com`) dari Footer.
- **Deskripsi Perubahan Spesifik**:
  1. Di `frontend/components/landing/navbar.tsx`:
     - Ganti menu navigasi menjadi `#games` (`Game Didukung`), `#features` (`Format & Kapabilitas`), dan `#how-it-works` (`Alur Kerja`).
     - Tambahkan state `mobileMenuOpen` dan tombol toggle hamburger (`Menu` / `X`) untuk layar `< 768px` sehingga pengguna mobile dapat mengakses seluruh navigasi dan tombol autentikasi.
     - Gunakan tombol CTA `bg-accent hover:bg-accent-hover text-white` untuk aksi utama (`Buat Turnamen` / `Dashboard`).
  2. Di `frontend/components/landing/footer.tsx`:
     - **Hapus link `<a href="https://github.com">GitHub</a>`** dan teks `"Modern Esports Tournament Engine"`.
     - Bangun layout footer 3 kolom yang proporsional:
       - Kolom 1: Brand Gridify & deskripsi singkat platform manajemen turnamen.
       - Kolom 2: 5 Format Kompetisi Resmi (`Single Elimination`, `Double Elimination`, `Round Robin League`, `Hybrid UCL`, `Battle Royale`).
       - Kolom 3: Navigasi Produk (`Game Didukung`, `Fitur Utama`, `Alur Kerja`, `Masuk Organizer`, `Daftar Akun`).
- **Acceptance Criteria**:
  - Tidak ada link ke `github.com` atau halaman fiktif di Footer maupun Navbar.
  - Navbar berfungsi penuh di desktop dan mobile (termasuk drawer menu mobile) serta menyesuaikan tombol aksi berdasarkan status login `user`.
- **Testing Requirements**:
  - Verifikasi seluruh link di Navbar dan Footer mengarah ke route atau anchor ID yang valid.

#### `UX-001` — Eliminasi Pola AI-Slop Global pada Landing Page, Champions Podium & Realtime Indicator
- **Status**: `DONE`
- **Prioritas**: `P1`
- **Dependencies**: `DS-001`, `NAV-001`, `HERO-002`
- **Lokasi Implementasi**:
  - `frontend/app/page.tsx`
  - `frontend/components/landing/features.tsx`
  - `frontend/components/landing/how-it-works.tsx`
  - `frontend/features/bracket/components/champions-view.tsx`
  - `frontend/components/providers/tournament/realtime-listener.tsx`
- **Alasan & Audit Finding**:
  Menjawab **CFM-08** dan **IMP-05**. Membersihkan repetisi kartu dekoratif, pill promosi generik, animasi Framer Motion yang berlebihan, dan floating pill SSE yang menghalangi sudut layar.
- **Deskripsi Perubahan Spesifik**:
  1. Di `frontend/components/landing/features.tsx`:
     - Ganti 6 kartu template seragam menjadi layout **Grid Kapabilitas Operasional** yang menyoroti 5 format kompetisi beserta aturan teknisnya (BO1/BO3/BO5, Home & Away, Cross-Seeding UCL, WWCD & Kill Points BR, serta QR Spectator Broadcast) tanpa menyebutkan istilah internal database (`PostgreSQL 17`, `owner_id`, `Go engine`).
  2. Di `frontend/components/landing/how-it-works.tsx`:
     - Ubah 3 kartu rata tengah menjadi **3-Step Operational Pipeline** rata kiri dengan indikator langkah monospace `STEP 01`–`STEP 03` dan detail alur kerja nyata.
  3. Di `frontend/app/page.tsx`:
     - Hapus pill `<Trophy /> SIAP MEMULAI KOMPETISI` dan ubah seksi CTA bawah menjadi *action banner* horizontal yang bersih.
  4. Di `frontend/features/bracket/components/champions-view.tsx`:
     - Hapus pill `"PODIUM JUARA TURNAMEN"` dan balok tinggi `h-28` bertuliskan angka `2, 1, 3`.
     - Desain ulang menjadi **Compact Championship Podium Banner** horizontal 3-kartu (Juara 1 dengan aksen emas `#F59E0B`, Juara 2 perak, Juara 3 perunggu) yang proporsional dan langsung menyatu di atas kanvas bracket/klasemen.
  5. Di `frontend/components/providers/tournament/realtime-listener.tsx`:
     - Tambahkan prop `inline?: boolean` (default `true` ketika dipasang di header, atau render badge status tipis yang terintegrasi di bar atas halaman bracket dan halaman publik penonton) sehingga tidak lagi melayang menutupi sudut kanan bawah layar.
- **Acceptance Criteria**:
  - Seluruh elemen AI-slop yang teridentifikasi pada **Section 6.1** telah bersih dari codebase.
  - `ChampionsView` tampil proporsional di semua format kompetisi tanpa memakan ruang vertikal berlebihan.
  - Indikator status koneksi `LIVE` terintegrasi rapi tanpa menutupi konten halaman.
- **Testing Requirements**:
  - Audit visual pada halaman `/`, `/dashboard/tournaments/[id]/bracket`, dan `/t/[id]`.

### Phase 5 — Global Page-by-Page UI/UX Consistency & Polish (`UX-002` & `UX-003`)
*Menyelaraskan seluruh halaman autentikasi, dashboard, pembuatan turnamen, manajemen peserta, dan halaman penonton publik.*

#### `UX-002` — Penyelarasan Halaman Auth (`login`, `register`, `forgot-password`, `update-password`) & Dashboard Shell
- **Status**: `DONE`
- **Prioritas**: `P1`
- **Dependencies**: `DS-001`, `DS-002`, `DS-003`
- **Lokasi Implementasi**:
  - `frontend/app/login/page.tsx`
  - `frontend/app/register/page.tsx`
  - `frontend/app/forgot-password/page.tsx`
  - `frontend/app/update-password/page.tsx`
  - `frontend/app/dashboard/layout.tsx`
  - `frontend/app/dashboard/page.tsx`
  - `frontend/app/dashboard/tournaments/page.tsx`
  - `frontend/app/dashboard/profile/page.tsx`
- **Alasan & Audit Finding**:
  Menjawab **Section 8 (Global UI/UX Consistency Audit)**. Memastikan seluruh halaman autentikasi dan dashboard konsisten dengan identitas visual baru, mengintegrasikan `GameLogoBadge` pada daftar turnamen, dan memperbaiki ergonomi tombol aksi di layar sentuh.
- **Deskripsi Perubahan Spesifik**:
  1. Pada `login/page.tsx`, `register/page.tsx`, `forgot-password/page.tsx`, dan `update-password/page.tsx`:
     - Gunakan wrapper input password yang konsisten sehingga tombol toggle mata (`Eye`/`EyeOff`) selalu berada tepat di tengah vertikal field input (`top-1/2 -translate-y-1/2`).
     - Terapkan `font-display` pada judul halaman dan identitas brand Gridify.
  2. Pada `dashboard/layout.tsx`:
     - Perjelas indikator navigasi aktif di sidebar dengan garis aksen kiri `border-l-2 border-accent bg-surface-elevated text-text-primary font-semibold`.
  3. Pada `dashboard/page.tsx` dan `dashboard/tournaments/page.tsx`:
     - Tampilkan `GameLogoBadge` (`size="sm"`) di samping nama game pada setiap kartu turnamen.
     - Ubah visibilitas `DeleteTournamentButton` agar pada layar mobile/touch tetap dapat diakses (`opacity-100 sm:opacity-0 sm:group-hover:opacity-100`).
     - Tambahkan tombol submit `"Cari"` eksplisit di dalam bar pencarian `dashboard/tournaments/page.tsx` agar pengguna mobile maupun desktop dapat mengeksekusi pencarian dengan mudah.
- **Acceptance Criteria**:
  - Halaman login, register, forgot-password, dan update-password tampil seragam dan rapi.
  - Kartu turnamen di Dashboard dan Turnamen Saya menampilkan logo dan nama game yang benar.
- **Testing Requirements**:
  - Uji navigasi sidebar, pencarian turnamen, dan responsivitas kartu di mobile & desktop.

#### `UX-003` — Redesign Workspace Turnamen (`[id]/*`), Pemilih Game Visual di Create Form & Halaman Publik Penonton (`/t/[id]`)
- **Status**: `DONE`
- **Prioritas**: `P1`
- **Dependencies**: `BUG-002`, `BUG-003`, `BUG-004`, `HERO-001`, `UX-001`
- **Lokasi Implementasi**:
  - `frontend/features/tournaments/components/create-tournament-form.tsx`
  - `frontend/app/dashboard/tournaments/[id]/layout.tsx`
  - `frontend/app/dashboard/tournaments/[id]/page.tsx`
  - `frontend/app/dashboard/tournaments/[id]/bracket/page.tsx`
  - `frontend/app/dashboard/tournaments/[id]/participants/page.tsx`
  - `frontend/app/(public)/t/[id]/page.tsx`
- **Alasan & Audit Finding**:
  Menjawab **Section 6, 7, dan 8**. Mengintegrasikan logo game pada alur pemilihan game di `CreateTournamentForm`, memperbaiki header workspace turnamen, dan merombak header halaman penonton `/t/[id]` agar fokus pada penyajian bagan/klasemen.
- **Deskripsi Perubahan Spesifik**:
  1. Di `frontend/features/tournaments/components/create-tournament-form.tsx`:
     - Ganti atau lengkapi pemilihan game dengan **Interactive Game Selector Grid** yang menampilkan `GameLogoBadge`, nama game, badge genre (`MOBA`, `FPS`, `SPORTS`, `BATTLE_ROYALE`), dan platform untuk ke-10 game, dengan state terpilih yang jelas (`border-2 border-accent bg-accent/12`).
  2. Di `frontend/app/dashboard/tournaments/[id]/layout.tsx` dan `page.tsx`:
     - Tampilkan `GameLogoBadge` di header detail turnamen bersama nama game (`tournament.game?.name`) dan format kompetisi.
  3. Di `frontend/app/(public)/t/[id]/page.tsx`:
     - Ganti Hero Banner tengah yang terlalu memakan tempat (dan hapus pill `"SPECTATOR LIVE VIEW"`) menjadi **Compact Broadcast Header** yang menampilkan `GameLogoBadge`, judul turnamen, status badge, jumlah peserta, tanggal, indikator `LIVE SSE` inline, dan tombol `Bagikan` (QR Code & Copy Link).
     - Periksa apakah pengguna yang membuka `/t/[id]` sedang login (`apiClient.users.getMe()`): jika ya, tampilkan tombol `"Dashboard Organizer"` (`/dashboard/tournaments/${id}`), jika tidak, tampilkan `"Masuk Organizer"` (`/login`).
- **Acceptance Criteria**:
  - Pemilihan game saat membuat turnamen menampilkan logo dan genre game secara visual dan langsung memfilter format kompetisi yang relevan.
  - Halaman publik `/t/[id]` langsung menampilkan bagan/klasemen di paruh atas layar tanpa terhalang banner raksasa atau floating pill.
- **Testing Requirements**:
  - Verifikasi alur pembuatan turnamen dan tampilan `/t/[id]` untuk seluruh 5 format kompetisi.

### Phase 6 — End-to-End Validation, Regression Testing & Build Verification (`TEST-*`)
*Memastikan kualitas akhir kode, kelulusan seluruh unit/integration test Go, dan kelulusan strict build Next.js.*

#### `TEST-001` — Verifikasi Unit & Integration Tests Backend Go + Kontrak DTO Standings/BR/Tournament
- **Status**: `DONE`
- **Prioritas**: `P0`
- **Dependencies**: `BUG-001`, `BUG-002`, `BUG-004`
- **Lokasi Implementasi**:
  - `backend/internal/service/standings_service_test.go`
  - `backend/internal/transport/http/handler/tournament_test.go`
  - `backend/internal/transport/http/handler/bracket_match_test.go`
- **Alasan & Audit Finding**:
  Memastikan penambahan field `participant_id`, `participant_name`, `br_results`, dan dukungan payload pembuatan turnamen teruji secara otomatis tanpa regresi.
- **Deskripsi Perubahan Spesifik**:
  1. Tambahkan asersi di `standings_service_test.go` bahwa setiap `StandingRow` dan `BRLeaderboardRow` yang dihasilkan memiliki `ParticipantID == Participant.ID` (bukan `uuid.Nil`) dan `ParticipantName == Participant.Name` (tidak kosong).
  2. Tambahkan pengujian di `tournament_test.go` untuk memverifikasi `POST /api/v1/tournaments` baik dengan field tingkat atas (`best_of`, `home_away`, `has_third_place`, `br_total_matches`) maupun dengan objek `match_rules`.
- **Acceptance Criteria**:
  - `go test -v -race ./...` di direktori `backend/` lulus 100% (0 kegagalan).
- **Testing Requirements**:
  - Eksekusi `cd backend && go test -v ./...`.

#### `TEST-002` — Verifikasi TypeScript Strict Build, Linting & Audit Console Error / Visual Regression di Frontend
- **Status**: `DONE`
- **Prioritas**: `P0`
- **Dependencies**: `DS-003`, `BRACKET-002`, `UX-002`, `UX-003`, `TEST-001`
- **Lokasi Implementasi**:
  - `frontend/`
- **Alasan & Audit Finding**:
  Memastikan seluruh perubahan di `frontend/` lulus pemeriksaan tipe TypeScript ketat (`ignoreBuildErrors: false`), bebas dari React console warnings, dan memenuhi seluruh kriteria penerimaan produk.
- **Deskripsi Perubahan Spesifik**:
  1. Jalankan `npx tsc --noEmit` dan `npm run build` di `frontend/`.
  2. Verifikasi tidak ada lagi referensi ke variabel CSS yang tidak terdefinisi (`--color-border-default` lama) atau properti DTO yang salah (`row.participant_id` tanpa fallback `row.participant?.id`, `match.participant_a_score` tanpa `match.score_a`, `tournament.games` tanpa `tournament.game`).
- **Acceptance Criteria**:
  - `npx tsc --noEmit` dan `npm run build` selesai dengan exit code `0` dan 0 warning/error.
  - Seluruh halaman publik dan dashboard bebas dari React Unique Key warning di console browser.
- **Testing Requirements**:
  - Eksekusi `cd frontend && npx tsc --noEmit && npm run build`.

## 12. Testing & Validation Strategy

1. **Static & Type-Safety Verification (`frontend/`)**:
   - Menjalankan `npx tsc --noEmit` dengan opsi `strict: true` dan `ignoreBuildErrors: false` di `next.config.ts`.
   - Memastikan seluruh komponen yang mengonsumsi `StandingRow`, `BRLeaderboardRow`, `Match`, `Tournament`, dan `Game` selaras dengan struktur JSON dari Go API.
2. **Backend Automated Test Suite (`backend/`)**:
   - Menjalankan seluruh suite pengujian Go (`go test -v ./internal/...`) yang mencakup:
     - `internal/service/bracket/*_test.go`: Pengujian matematika seeding, Single Elimination (termasuk BYE & Bronze Match), Double Elimination (termasuk BYE & drop-down LB), Round Robin (1-leg & 2-leg), dan Battle Royale.
     - `internal/service/standings_service_test.go`: Pengujian kalkulasi klasemen, leaderboard BR, podium juara, serta keberadaan field `participant_id` dan `participant_name`.
     - `internal/transport/http/handler/*_test.go`: Pengujian autentikasi, otorisasi kepemilikan, pembuatan turnamen, dan pembaruan skor.
3. **Manual & Visual Verification Scenarios**:
   - **Skenario 1 (StandingsTable & Hybrid UCL)**: Buka turnamen Round Robin & Hybrid UCL → pastikan tidak ada error `Each child in a list should have a unique "key" prop` di console, seluruh nama tim muncul, dan 2 tim teratas di Grup A & Grup B memiliki badge `Q`.
   - **Skenario 2 (Create Tournament & Format Selection)**: Buka `/dashboard/create` → pilih salah satu dari 10 game → klik masing-masing format kompetisi → pastikan kartu yang dipilih menampilkan border Crimson Valor `2px`, latar ter-highlight, dan indikator centang aktif → klik `"Buat Turnamen Sekarang"` → pastikan berhasil diarahkan ke halaman detail turnamen baru (`201 Created`).
   - **Skenario 3 (Bracket Connectors & Progression)**: Buka turnamen Single Elimination (dengan Bronze Match) dan Double Elimination → pastikan Upper Bracket, Lower Bracket, Grand Final, dan Bronze Match berada di seksi yang benar → pastikan garis konektor SVG terlihat jelas dan menunjuk tepat ke baris Slot A (atas) atau Slot B (bawah) pada kartu match tujuan.
   - **Skenario 4 (Hero, Navbar & Footer)**: Buka `/` pada viewport desktop dan mobile → pastikan badge `"Arsitektur Baru: Next.js 16 + Go + Realtime SSE"` sudah tidak ada → pastikan 10 logo game resmi tampil rapi dan interaktif → pastikan menu mobile berfungsi dan link `GitHub` sudah dihapus dari Footer.

## 13. Risk Register

| ID Risiko | Deskripsi Risiko | Dampak | Probabilitas | Strategi Mitigasi |
| :--- | :--- | :--- | :--- | :--- |
| **RISK-UI-01** | **Perbedaan posisi koordinat SVG saat font web selesai dimuat (*Font Loading Layout Shift*)** yang dapat membuat titik ujung konektor bracket bergeser 2–4px. | Sedang | Sedang | Gunakan `ResizeObserver` pada kontainer kanvas bracket dan jalankan ulang `calculatePaths` setelah `document.fonts.ready` selesai. |
| **RISK-UI-02** | **Kepadatan visual kanvas bracket pada Double Elimination dengan 16+ peserta di layar kecil**. | Sedang | Tinggi | Pertahankan lebar kartu match yang kompak (`w-60` / `240px`), gunakan `overflow-x-auto` dengan scrollbar kustom yang halus, dan pisahkan seksi Upper Bracket & Lower Bracket dengan header berlabel jelas. |
| **RISK-UI-03** | **Kegagalan pemuatan aset logo jika ke depan ditambahkan game baru di tabel `games` tanpa file SVG lokal**. | Rendah | Sedang | Komponen `GameLogoBadge` dilengkapi state `hasError` (`onError`) yang secara otomatis merender emblem vektor berbasis genre (`MOBA`, `FPS`, `SPORTS`, `BATTLE_ROYALE`) + inisial game. |
| **RISK-API-01** | **Ketidakcocokan format payload antara form frontend dan `DisallowUnknownFields()` di handler Go**. | Tinggi | Rendah | Selaraskan payload di `create-tournament-form.tsx` agar mengirim field tingkat atas yang sesuai dengan `CreateTournamentPayload`, sekaligus tambahkan dukungan field opsional `match_rules` di struct Go. |

## 14. Definition of Done (DoD)

Seluruh pekerjaan pada rencana implementasi ini dinyatakan **SELESAI (DONE)** hanya apabila seluruh kriteria berikut terpenuhi:
1. **Zero Console Errors**: Tidak ada lagi error React `Each child in a list should have a unique "key" prop` pada `StandingsTable`, `BRLeaderboard`, maupun komponen list lainnya.
2. **Clear Selection Feedback & Working Form Submission**: Pilihan Game, Format Kompetisi, Best-Of, Putaran Battle Royale, dan Status Turnamen memiliki feedback visual aktif yang jelas, mendukung navigasi keyboard, dan form pembuatan turnamen berhasil disimpan ke backend Go tanpa error `400`.
3. **Accurate Score & Game Name Rendering**: Seluruh kartu pertandingan (`MatchCard`, `LeagueMatchCard`), modal input skor (`ScoreModal`, `BRScoreModal`), dan header/kartu turnamen menampilkan skor (`score_a`, `score_b`) dan nama game (`tournament.game.name`) yang akurat.
4. **Cohesive Visual Identity & Typography**: Sistem warna **"Tactical Broadcast Slate & Crimson Volt"** (`@theme` di `globals.css`) dan sistem tipografi 3-tier (`Space Grotesk`, `Plus Jakarta Sans`, `JetBrains Mono`) aktif secara konsisten di seluruh halaman.
5. **Zero AI-Slop**: Badge tech stack di Hero section, pill dekoratif berulang, link `GitHub` di Footer, link palsu `"Bagan Publik"` di Navbar, dan floating pill `SSE REALTIME` di sudut kanan bawah telah dihapus/dirombak sesuai spesifikasi.
6. **Accurate Bracket Connectors**: Bagan Single Elimination (termasuk Bronze Match) dan Double Elimination (Upper, Lower, Grand Final) dikelompokkan berdasarkan `bracket_type` dengan benar, dan garis konektor SVG ortogonal menghubungkan setiap pertandingan ke baris slot peserta (`A` atau `B`) yang tepat.
7. **Supported Game Logos in Hero & App**: Ke-10 game resmi di database memiliki aset logo SVG di `frontend/public/games/` dengan fallback yang aman, ditampilkan secara art-directed di Hero section serta diintegrasikan pada alur pemilihan game dan kartu turnamen.
8. **Green Build & Tests**: `go test ./...` di `backend/` dan `npx tsc --noEmit && npm run build` di `frontend/` lulus 100% tanpa error.

## 15. Execution Checklist

- [x] **`DS-001`** (`P0`) — Implementasi Tailwind CSS v4 `@theme` & Palet Warna "Tactical Broadcast Slate & Crimson Volt" di `frontend/app/globals.css`
- [x] **`DS-002`** (`P1`) — Implementasi Sistem Tipografi 3-Tier (`Space Grotesk`, `Plus Jakarta Sans`, `JetBrains Mono`) di `frontend/app/layout.tsx` & `globals.css`
- [x] **`DS-003`** (`P1`) — Penyelarasan Komponen UI Dasar (`Button`, `Input`, `Select`, `Card`, `Badge`, `Modal`) di `frontend/components/ui/*`
- [x] **`BUG-001`** (`P0`) — Fix React Unique Key Warning, Blank Team Names & Group Qualification Highlight pada `StandingsTable` & `standings_service.go`
- [x] **`BUG-002`** (`P0`) — Fix Visual Selection Feedback, Keyboard Accessibility & Payload Submission pada `CreateTournamentForm` & `TournamentSettingsForm`
- [x] **`BUG-003`** (`P0`) — Fix Score Display & Input Binding (`score_a` / `score_b`) serta Game Name Resolution (`tournament.game`)
- [x] **`BUG-004`** (`P0`) — Fix Battle Royale Leaderboard, Round Numbering (`round_number`), Playing Groups & Match Results Pre-population
- [x] **`BRACKET-001`** (`P0`) — Implementasi Pemisahan `bracket_type` & Layout Presisi untuk Single/Double Elimination, Grand Final & Bronze Match
- [x] **`BRACKET-002`** (`P0`) — Implementasi Mesin Konektor SVG Ortogonal Berbasis Slot (`next_match_slot`) & `ResizeObserver`
- [x] **`HERO-001`** (`P1`) — Pembuatan Aset Logo Vektor 10 Game Resmi (`frontend/public/games/*.svg`) & Komponen `GameLogoBadge` dengan Fallback
- [x] **`HERO-002`** (`P1`) — Redesign Hero Section & Interactive Bracket Preview dengan Roster 10 Game Kompetitif (Hapus Badge Tech Stack)
- [x] **`NAV-001`** (`P1`) — Redesign Navbar (Tambah Mobile Drawer & Perbaiki Link) & Footer Produk (Hapus Link GitHub)
- [x] **`UX-001`** (`P1`) — Eliminasi Pola AI-Slop Global pada Landing Page (`Features`, `HowItWorks`, `CTA`), `ChampionsView` & `RealtimeListener`
- [x] **`UX-002`** (`P1`) — Penyelarasan Halaman Auth (`login`, `register`, `forgot-password`, `update-password`) & Dashboard Shell/Cards
- [x] **`UX-003`** (`P1`) — Redesign Workspace Turnamen (`[id]/*`), Pemilih Game Visual di Create Form & Halaman Publik Penonton (`/t/[id]`)
- [x] **`TEST-001`** (`P0`) — Verifikasi Unit & Integration Tests Backend Go (`go test ./...`)
- [x] **`TEST-002`** (`P0`) — Verifikasi TypeScript Strict Build (`tsc --noEmit` & `next build`) dan Audit Regresi Visual

