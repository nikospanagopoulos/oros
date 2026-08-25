# Master PRD — Oros

**Τύπος εγγράφου:** Master PRD (συνολική εικόνα project, όλες οι εκδόσεις)
**Ημερομηνία:** 2026-07-12
**Ιδιοκτήτης:** Νίκος
**Σχετικά έγγραφα:** `docs/adr/` για αναλυτικό ιστορικό τεχνικών αποφάσεων

**Σχετικά με το όνομα:** Το "Oros" (Όρος) επιλέχθηκε μετά από έρευνα
που έδειξε ότι σχεδόν κάθε προφανές αγγλικό ή ελληνικό-μυθολογικό όνομα
στο χώρο "AI agent security/guardrails" (AgentShield, AgentGuard,
Sentinel, Aegis, Watchtower, Veto, Janus, Hecate, Talos, Warden, Reins,
Θέμις, Ευνομία) ήταν ήδη κατειλημμένο — ένδειξη πόσο γρήγορα κινείται
αυτή η αγορά το 2026. Το "Όρος" έχει διπλό νόημα στα ελληνικά: (α) η
αρχαία πέτρα-σύνορο (ὅρος) που σηματοδοτούσε όριο ιδιοκτησίας, και (β)
ο σύγχρονος "όρος/κανόνας" (term/condition) — και τα δύο ταιριάζουν
άμεσα με το ρόλο του εργαλείου: καταγράφει όρια (Recorder) και επιβάλλει
κανόνες (Protection).

---

## 0. North Star

Το προϊόν υπάρχει για να δίνει στους developers πλήρη ορατότητα και
έλεγχο πάνω στις ενέργειες των AI coding agents, με τρόπο προβλέψιμο,
ασφαλή και απλό.

Αν ένα νέο feature δεν ενισχύει αυτή την κατεύθυνση, πιθανότατα δεν
ανήκει στο προϊόν.

---

## 1. Executive Summary

Οι AI coding agents (Claude Code, Cursor κ.ά.) εκτελούν πλέον εντολές,
τροποποιούν αρχεία και κάνουν commits αυτόνομα, αλλά κανένα εργαλείο δεν
καταγράφει συστηματικά *τι* έκαναν, *γιατί* το έκαναν, και δεν τους
σταματά *πριν* κάνουν κάτι επικίνδυνο. Το Oros είναι ένα
local-first CLI εργαλείο που κάθεται ανάμεσα στον developer και τον AI
agent: καταγράφει κάθε ενέργεια (Recorder), την οπτικοποιεί (Timeline &
Replay), και τελικά μπλοκάρει ή ζητά έγκριση για επικίνδυνες ενέργειες
(Policies). Διαφέρει από τα υπάρχοντα observability εργαλεία (AgentOps,
Langfuse, LangSmith) γιατί δουλεύει στο επίπεδο του developer machine
(process/OS level), όχι μέσα από SDK instrumentation σε production
εφαρμογές, και γιατί προσφέρει enforcement πριν την εκτέλεση, όχι μόνο
παρατήρηση μετά.

---

## 2. Vision & Positioning

**Vision:** Κάθε developer που χρησιμοποιεί AI agents να έχει την ίδια
σιγουριά που έχει με ένα CI pipeline: ξέρει τι έτρεξε, μπορεί να το
ξαναδεί, και εμπιστεύεται ότι δεν θα συμβεί κάτι καταστροφικό χωρίς
έλεγχο.

**Mission:** Να χτίσουμε το πρώτο local-first εργαλείο που δίνει
recording, replay και enforcement πάνω σε AI coding agents, χωρίς
πολύπλοκη εγκατάσταση.

**Long-term goal:** Από απλό CLI εργαλείο ενός developer, σε πλατφόρμα
που ομάδες και εταιρείες χρησιμοποιούν για να διέπουν με ασφάλεια τη
χρήση AI agents στην ανάπτυξη λογισμικού τους.

**Positioning (wedge):** Δεν μπαίνουμε στην αγορά ως "AI observability
platform" (κατειλημμένη, πυκνή αγορά με funded παίκτες όπως Langfuse,
LangSmith, AgentOps, Braintrust). Μπαίνουμε ως **security/enforcement
εργαλείο** — "μην αφήσεις τον agent σου να καταστρέψει το repo σου" —
μια πρόταση που ένας developer καταλαβαίνει σε 3 δευτερόλεπτα. Το
ευρύτερο platform vision (analytics, compliance, multi-agent) χτίζεται
πάνω σε αυτό το wedge, όχι πριν από αυτό.

---

## 3. Problem Statement

Οι developers που χρησιμοποιούν AI coding agents αντιμετωπίζουν τρία
συνδεδεμένα προβλήματα:

1. **Δεν ξέρουν τι έκανε ο agent.** Καμία συστηματική καταγραφή
   εντολών, αρχείων που άλλαξαν, ή αποτελεσμάτων.
2. **Δεν καταλαβαίνουν γιατί το έκανε.** Όταν κάτι σπάει, δεν υπάρχει
   εύκολος τρόπος να δουν την αλυσίδα ενεργειών που οδήγησε εκεί.
3. **Δεν μπορούν να τον σταματήσουν πριν κάνει ζημιά.** Καμία επιβολή
   κανόνων πριν την εκτέλεση μιας επικίνδυνης ενέργειας (π.χ. force
   push σε main, διαγραφή migrations, έκθεση secrets).

Αυτό δεν είναι υποθετικό πρόβλημα — το 72% των security alerts στα
υπάρχοντα εργαλεία θεωρείται θόρυβος αντί για προστασία (JFrog, 2026),
κάτι που δείχνει ότι η ανάγκη για ένα καθαρό, στοχευμένο εργαλείο με
λίγους, σωστούς κανόνες —όχι ένα ακόμα θορυβώδες dashboard— είναι
πραγματική.

---

## 4. Competitive Landscape

| Εργαλείο | Δυνατό σημείο | Αδυναμία |
|---|---|---|
| AgentOps | Session replay, time-travel debugging, 400+ frameworks | SDK-based (μέσα στην εφαρμογή), όχι OS/process level· καμία επιβολή πριν την εκτέλεση |
| Langfuse | Open-source, self-hosted, τεράστια υιοθέτηση | Ίδιο: production LLM observability, όχι developer-machine enforcement |
| LangSmith | Πλήρες lifecycle observability, ελάχιστο overhead | Ίδιο μοτίβο: παρατηρεί, δεν μπλοκάρει |
| Braintrust | Ισχυρό CI/CD eval-gating, γενναιόδωρο free tier | Εστιάζει σε evals/regressions, όχι σε real-time guardrails για local agents |
| Earthly (πρώην) | — | Έκλεισε το CI προϊόν το 2025, έκανε pivot σε AI guardrails (Lunar) — ένδειξη ότι η αγορά κατευθύνεται εδώ, αλλά ακόμα δεν υπάρχει ώριμος παίκτης στο συγκεκριμένο wedge |

**Πού διαφοροποιούμαστε:** Κανένα από τα παραπάνω δεν λειτουργεί στο
επίπεδο του developer's shell/process πριν την εκτέλεση μιας εντολής.
Όλα είναι SDK-instrumentation μέσα σε production εφαρμογές. Το κενό
είναι: local-first, process-level, με πραγματικό pre-flight enforcement
(όχι μόνο παρατήρηση μετά τα γεγονότα).

---

## 5. Target Users

**Πρωταρχικός χρήστης (v0.1–v0.3):** Solo developer ή μικρή ομάδα (2-10
άτομα) που χρησιμοποιεί καθημερινά CLI-based AI coding agents (Claude
Code, Aider) και θέλει έλεγχο/καταγραφή χωρίς πολύπλοκο setup. (Χρήστες
IDE-based agents όπως ο Cursor γίνονται target μετά το v1.0 — βλ.
Scope, ενότητα 6.)

**Δευτερεύων χρήστης (v0.4+):** Engineering lead ή CTO μικρής/μεσαίας
startup που χρειάζεται ορατότητα σε επίπεδο ομάδας — ποιος χρησιμοποιεί
AI agents, τι έκαναν, αν τηρούνται policies.

**Μακροπρόθεσμος χρήστης (v1.0+):** Compliance/security officer σε
εταιρεία που πρέπει να αποδείξει πώς ελέγχονται τα AI εργαλεία στο
development pipeline (SOC2-style ανάγκες).

---

## 6. Scope

**Τι κάνει:**
- Καταγράφει εντολές, αλλαγές αρχείων και exit codes AI agents που
  τρέχουν μέσα από το εργαλείο (process wrapping).
- Οπτικοποιεί το ιστορικό ενεργειών σε timeline, με δυνατότητα replay.
- Επιβάλλει κανόνες (policies) πριν την εκτέλεση επικίνδυνων ενεργειών.
- Παρέχει AI εξήγηση ("γιατί απέτυχε;") βασισμένη στο πραγματικό
  ιστορικό ενεργειών.

**Τι ΔΕΝ κάνει (τουλάχιστον όχι στο MVP):**
- Δεν υποστηρίζει IDE-based agents (Cursor, VS Code Copilot Edits) στο
  v0.1–v0.3 — μόνο CLI-based agents που ξεκινούν ρητά μέσω terminal
  (Claude Code, Aider). Το process-wrapping μοντέλο δεν λειτουργεί για
  agents που τρέχουν ως background binaries μέσα σε IDE. Υποστήριξη για
  IDE-based agents μετατίθεται για μετά το v1.0 (πιθανώς μέσω custom
  editor extension).
- Δεν κάνει kernel-level / πραγματικό OS-level interception (είναι
  process-level wrapping, βλ. ADR).
- Δεν κάνει static analysis πολλαπλών γλωσσών ή architecture graphs
  (μετατέθηκε μετά το v1.0).
- Δεν δίνει πιθανοτικά "risk scores" (π.χ. "85% επικίνδυνο") — μόνο
  deterministic κανόνες + AI εξήγηση, ποτέ εφευρημένη ακρίβεια.
- Δεν προσφέρει multi-agent orchestration ή marketplace plugins στο
  MVP (long-term ιδέες, βλ. ενότητα 13).

---

## 7. Product & Design Principles

**Γενικές αρχές προϊόντος:**
- **Local-first** — δουλεύει πλήρως χωρίς cloud, χωρίς server setup.
- **Security-first** — κάθε απόφαση σχεδίασης προτεραιοποιεί την
  ασφάλεια των δεδομένων του χρήστη.
- **Explainable** — καμία "μαύρη κούτα"· κάθε απόφαση (block/allow)
  έχει ξεκάθαρη αιτιολόγηση.
- **Offline capable** — δεν απαιτεί σύνδεση internet για το core
  λειτουργικό (v0.1–v0.3).
- **Minimal dependencies** — λίγες, σταθερές εξωτερικές βιβλιοθήκες.
- **Fast startup** — το CLI πρέπει να ξεκινάει άμεσα, μηδενική
  αντιληπτή καθυστέρηση.
- **Privacy by design** — τα δεδομένα μένουν στο μηχάνημα του χρήστη
  εκτός αν ρητά επιλέξει cloud sync (v0.4+).

**Τεχνικές/ασφάλειας αρχές:**
- **Redaction πάντα ενεργό** — κανένα secret δεν αποθηκεύεται ποτέ σε
  raw μορφή στα logs (regex sanitization από το v0.1, βασισμένο σε
  patterns τύπου gitleaks/truffleHog).
- **Auditability** — κάθε ενέργεια αφήνει ίχνος, append-only.
- **Deterministic policies, όχι εφευρημένα risk scores** — οι κανόνες
  ασφαλείας είναι πάντα κανόνες matching (π.χ. "αλλαγή σε .env = HIGH
  RISK"), ποτέ πιθανοτική εκτίμηση από μοντέλο. Το AI χρησιμοποιείται
  μόνο στο Intelligence layer για εξηγήσεις, ποτέ για αποφάσεις
  ασφαλείας.
- **File permissions 0600 + `.gitignore`** για κάθε τοπικό αρχείο
  δεδομένων (SQLite database).
- **Ρητό Definition of Done ανά έκδοση** — κάθε version PRD (ξεκινώντας
  από το v0.1) πρέπει να περιέχει συγκεκριμένο, ελέγξιμο checklist
  ("done" σημαίνει: X δουλεύει, tests περνάνε, README ολοκληρώθηκε,
  κ.λπ.). Το checklist ζει στο PRD της εκάστοτε έκδοσης, όχι εδώ.

---

## 8. High-Level Architecture

Το προϊόν οργανώνεται σε 3 επίπεδα:

```
Oros
│
├── Core Engine        (v0.1 – v0.2)
│   ├── Recorder
│   ├── Timeline
│   └── Replay
│
├── Protection          (v0.3)
│   ├── Policies
│   ├── Guardrails
│   └── Deterministic Pre-flight Checks
│
└── Intelligence        (v0.5+)
    ├── AI Analysis ("γιατί απέτυχε;")
    └── (μελλοντικά: Living Docs, Architecture Graph — μετά v1.0)
```

**Architecture Principles** (πιο θεμελιώδεις από τα γενικά product
principles της ενότητας 7 — αφορούν το πώς χτίζεται ο κώδικας):
- **Event First** — κάθε ενέργεια αναπαρίσταται πρώτα ως event, πριν
  γίνει οτιδήποτε άλλο (log, block, εξήγηση).
- **Local First** — ο πυρήνας δεν εξαρτάται ποτέ από δίκτυο/cloud.
- **Security First** — καμία απόφαση αρχιτεκτονικής δεν θυσιάζει
  ασφάλεια για ευκολία.
- **Simple before Flexible** — hardcoded/απλές λύσεις προτιμώνται έναντι
  γενικευμένων μηχανισμών (π.χ. plugin system) μέχρι να αποδειχθεί
  πραγματική ανάγκη (βλ. ADR-001).
- **Backward Compatible Events** — το event schema εξελίσσεται χωρίς να
  σπάει παλιά δεδομένα (βλ. `SchemaVersion` παρακάτω).

**Architecture pattern:** Event-driven core με απλά Go interfaces (όχι
plugin system στο MVP· βλ. ADR-001). Ο Recorder παράγει structured
events με typed metadata (`EventType`, `Status`, `Severity` — όλα typed
enums, όχι plain strings), globally unique `ID` (UUIDv7, sortable by
δημιουργία), ρητά ορισμένο `Timestamp` (Unix milliseconds, UTC), και
γενικό `Payload any` για το ελεύθερο περιεχόμενο. Τα events γράφονται
στη SQLite, και κάθε νέο layer (Protection, Intelligence) υλοποιεί ένα
κοινό `EventHandler` interface για να τα «ακούει», χωρίς να ξαναγράφεται
ο πυρήνας. Το πεδίο `Status`
(`EXECUTED`/`PENDING_APPROVAL`/`BLOCKED`) διασφαλίζει ότι το audit log
παραμένει ακριβές όταν μια ενέργεια περιμένει έγκριση χρήστη (v0.3).
Αναλυτικά στο ADR-001.

**Interaction model:** Ο AI agent τρέχει *μέσα* από το CLI wrapper
(π.χ. `oros -- claude`), όχι δίπλα του — αποφεύγεται η εύθραυστη
ανίχνευση μέσω process tree/PPID matching (βλ. ADR για λεπτομέρειες).
**Περιορισμός:** αυτό το μοντέλο δουλεύει μόνο για CLI-based agents
(Claude Code, Aider) που ξεκινούν ρητά από το terminal. IDE-based agents
που τρέχουν background binaries (π.χ. Cursor, VS Code Copilot Edits)
**δεν** υποστηρίζονται στο v0.1–v0.3 — βλ. Scope, ενότητα 6.

---

## 9. Technical Stack

- **Γλώσσα:** Go (CLI + αργότερα native web dashboard, χωρίς δεύτερο
  stack στο MVP)
- **Αποθήκευση:** SQLite (`mattn/go-sqlite3` ή `modernc.org/sqlite`),
  τοπικό αρχείο, append-only events
- **File watching:** `fsnotify`
- **CLI framework:** `cobra` ή `urfave/cli`
- **Terminal UI (v0.2+):** `bubbletea`
- **AI integration (v0.5+):** Anthropic API
- **Policy format (v0.3):** Go structs + YAML rules
- **Web dashboard (v0.4+):** Go-native (`net/http` + `html/template`)
  σε συνδυασμό με **HTMX** για live updates/interactivity χωρίς
  JavaScript framework — δίνει αίσθηση SPA γράφοντας μόνο Go· αποφεύγεται
  δεύτερο τεχνολογικό stack (Next.js/React)

*(Οι τεχνολογίες μπορεί να αλλάξουν· βλ. `docs/adr/` για αναλυτικές
αποφάσεις ανά περίπτωση.)*

---

## 10. Product Roadmap

| Έκδοση | Επίπεδο | Περιεχόμενο |
|---|---|---|
| v0.1 | Core Engine | Recorder — καταγραφή εντολών, timestamps, exit codes, redaction, audit log |
| v0.2 | Core Engine | Timeline & Replay (TUI με bubbletea), search, filters |
| v0.3 | Protection | Policies/Guardrails — deterministic πρόληψη επικίνδυνων ενεργειών, human approval |
| v0.4 | Platform/SaaS | Web Dashboard (Go-native + HTMX), cloud sync, team policies |
| v0.5 | Intelligence | AI Analysis — "γιατί απέτυχε;" μέσω Anthropic API πάνω στο τοπικό timeline |
| v1.0 | Enterprise | Teams, SSO, API, compliance features |

Κάθε έκδοση είναι ανεξάρτητα χρήσιμη και shippable· δεν εξαρτάται από
την ολοκλήρωση της επόμενης.

---

## 11. Long-term Vision (3–5 χρόνια)

Πέρα από το v1.0, πιθανές κατευθύνσεις (όχι δεσμεύσεις):
- Multi-agent orchestration & παρακολούθηση πολλών agents ταυτόχρονα
- Πλήρες enterprise governance & compliance layer (SOC2/ISO27001-style
  readiness)
- Marketplace για custom policies/plugins τρίτων
- Cloud edition (πλήρως hosted) παράλληλα με self-hosted edition

---

## 12. Business Model & Monetization (Open-Core Strategy)

**Free / Open-Source tier (v0.1 – v0.3):** Το Go binary, ο Recorder, το
TUI Timeline/Replay και το τοπικό Policy engine είναι δωρεάν και
ανοιχτού κώδικα. Στόχος: organic growth μέσω GitHub, Hacker News,
τεχνικών κοινοτήτων.

**Commercial tier (v0.4 – v1.0):** Χτίζεται πάνω στο δωρεάν core, όχι
αντί αυτού.
- **Team plan:** κεντρικό web dashboard, cloud αποθήκευση logs, κοινά
  policies για ομάδα. Τιμολόγηση: *προς επικύρωση* μετά από feedback
  πρώτων χρηστών (ενδεικτικό εύρος αγοράς για αντίστοιχα DevSecOps
  εργαλεία: $10–30/χρήστη/μήνα — δεν έχει επιβεβαιωθεί ακόμα).
- **Enterprise plan:** SSO/SAML, απεριόριστη διατήρηση logs, compliance
  reporting. Τιμολόγηση: custom, *προς καθορισμό*.

Η αρχή του open-core: αν χρεωθεί το CLI από την πρώτη μέρα, η
υιοθέτηση θα μειωθεί δραματικά. Το monetization "κουμπώνει" μόνο όταν
εμφανίζεται πραγματική ανάγκη ομάδας/εταιρείας (v0.4+).

---

## 13. Success Metrics

**Key Metric (γενικό, μακροπρόθεσμο):** Weekly Protected Sessions — ο
αριθμός development sessions ανά εβδομάδα που τρέχουν με ενεργό το
Oros.

**Product metrics (v0.1 – v0.3, local CLI):**
- Retention: πόσοι χρήστες που εγκατέστησαν το εργαλείο συνεχίζουν να
  το τρέχουν μετά από 2-4 εβδομάδες (target: *προς επικύρωση* μετά τα
  πρώτα πραγματικά δεδομένα).
- Πόσα επικίνδυνα commands μπλόκαρε το policy engine τοπικά (δείκτης
  πραγματικής αξίας, όχι μόνο χρήσης).

**Business metrics (v0.4 – v1.0):**
- Μετατροπή από free CLI σε paid team plan (target: *προς επικύρωση*).
- MRR growth μετά το launch του v0.4 (target: *προς επικύρωση*).

*Σημείωση: Δεν ορίζουμε συγκεκριμένα ποσοστά/νούμερα-στόχους πριν
έχουμε πραγματικά δεδομένα χρήσης — τα νούμερα θα οριστούν μετά τις
πρώτες εβδομάδες του v0.1-v0.3, όχι εκ των προτέρων ως εικασία.*

---

## 14. Risks

- **Ανταγωνισμός από funded παίκτες** (Langfuse/AgentOps κ.λπ.) αν
  αλλάξουν κατεύθυνση προς process-level enforcement.
- **Τεχνικό ρίσκο στο process-wrapping μοντέλο** — πρέπει να
  επιβεβαιωθεί μέσω τεχνικών spikes πριν το πλήρες v0.1 (π.χ. αν το
  wrapping δουλεύει αξιόπιστα σε macOS/Linux/Windows).
- **Χρονικός περιορισμός** — solo developer, 6 μήνες, ρίσκο
  υπερφόρτωσης αν το scope διευρυνθεί πρόωρα (π.χ. Living Docs πριν
  την ώρα του).
- **Adoption risk** — αν το free tier δεν προσφέρει αρκετή αυτόνομη
  αξία, δεν θα υπάρξει organic growth για να στηρίξει το μελλοντικό
  monetization.
- **False positives στο Policy engine** — αν οι κανόνες μπλοκάρουν
  υπερβολικά, θα εγκαταλειφθεί γρήγορα από τους χρήστες (72% θεωρούν
  ήδη τα security alerts στα υπάρχοντα εργαλεία θόρυβο).

---

## 15. Constraints & Assumptions

**Constraints:**
- Solo developer
- 6 μήνες για MVP (v0.1 – v0.3 τουλάχιστον)
- Laptop-only development environment
- Go ως βασική γλώσσα
- Local-first αρχιτεκτονική
- Open-source core

**Assumptions:**
- Η χρήση AI coding agents θα συνεχίσει να αυξάνεται.
- Οι developers προτιμούν τοπικά εργαλεία έναντι cloud-only λύσεων για
  ευαίσθητα δεδομένα κώδικα.
- Οι μικρές ομάδες δεν θέλουν πολύπλοκη εγκατάσταση/configuration.

*Αν κάποια από αυτές τις υποθέσεις αποδειχθεί λανθασμένη, αυτό είναι
σημάδι να επανεξεταστεί η αντίστοιχη ενότητα του PRD.*

---

## 16. Future Ideas (Backlog — όχι δεσμεύσεις)

- Slack notifications για policy violations
- GitHub Action integration
- Docker/Kubernetes-aware policies
- Living documentation / architecture graph (μετά v1.0)
- Πλήρες plugin system για τρίτους (μετά v1.0, βλ. ADR-001)

---

## 17. Glossary

- **Agent:** Ο AI coding agent (π.χ. Claude Code, Cursor) που εκτελεί
  ενέργειες μέσα στο development περιβάλλον.
- **Session:** Μία συνεδρία εκτέλεσης του agent μέσα από το wrapper.
- **Run:** Μία μεμονωμένη εκτέλεση εντολής/ενέργειας μέσα σε session.
- **Event:** Structured καταγραφή μίας ενέργειας (εντολή, αλλαγή
  αρχείου, κ.λπ.).
- **Recorder:** Το component που καταγράφει events.
- **Timeline:** Η οπτική αναπαράσταση της χρονολογικής σειράς events.
- **Replay:** Η δυνατότητα να ξαναδείς/περιηγηθείς σε ένα προηγούμενο
  session βήμα-βήμα.
- **Policy:** Κανόνας που καθορίζει τι επιτρέπεται/απαγορεύεται.
- **Guardrail:** Μηχανισμός επιβολής μίας policy σε πραγματικό χρόνο.
- **Audit log:** Το πλήρες, append-only ιστορικό events.
- **Pre-flight check:** Έλεγχος μιας ενέργειας πριν εκτελεστεί.