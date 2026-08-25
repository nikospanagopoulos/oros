# ADR-0001: Event-Driven-Core (όχι Plugin System) στο v0.1

**Status:** Accepted
**Date:** 2026-07-12

## Context

Το project χωρίζεται σε 3 αρχιτεκτονικά επίπεδα: Core Engine (Recorder,
Timeline, Replay), Protection (Policies, Guardrails) και Intelligence
(AI Analysis). Κάθε επίπεδο χτίζεται σε διαφορετική έκδοση (v0.1 → v0.5+),
οπότε ο πυρήνας πρέπει να επιτρέπει σε νέα layers να "ακούν" τα ίδια
δεδομένα χωρίς να ξαναγραφεί ο Recorder κάθε φορά.

Δύο βασικές επιλογές εξετάστηκαν:

1. **Plugin-based architecture** (σαν Terraform providers): δυναμική
   φόρτωση modules, ορισμένο plugin API/contract, versioning.
2. **Απλό event-driven μοντέλο**: ο Recorder παράγει structured events,
   άλλα modules υλοποιούν ένα κοινό interface και τα "ακούν".

## Decision

Επιλέγουμε **απλό event-driven μοντέλο με hardcoded handlers** για τις
εκδόσεις v0.1–v0.3. Όχι plugin system.

Ροή δεδομένων:

```
Shell / Agent
      │
      ▼
  Recorder
      │
      ▼
   Event
      │
      ▼
  SQLite  (γράφεται πρώτα, πριν σταλεί πουθενά)
      │
      ▼
 Dispatcher
      │
 ┌────┼─────┐
 ▼    ▼     ▼
Timeline  Policy  Dashboard
(v0.2)    (v0.3)  (v0.4+)
```

Βασικό σχήμα (typed metadata + generic payload — βλ. αναθεώρηση
2026-07-12 παρακάτω):

```go
type EventType string

const (
    CommandExecuted EventType = "command.executed"
    FileModified    EventType = "file.modified"
    FileCreated     EventType = "file.created"
    FileDeleted     EventType = "file.deleted"
)

type Severity string

const (
    SeverityInfo    Severity = "info"
    SeverityWarning Severity = "warning"
    SeverityHigh    Severity = "high"
)

type EventStatus string

const (
    StatusExecuted        EventStatus = "EXECUTED"
    StatusPendingApproval EventStatus = "PENDING_APPROVAL"
    StatusBlocked         EventStatus = "BLOCKED"
)

type Event struct {
    ID            string      // UUIDv7 (βλ. σημείωση παρακάτω)
    Timestamp     int64       // Unix milliseconds, UTC — ρητά ορισμένο, όχι διφορούμενο
    AgentID       string
    EventType     EventType
    Status        EventStatus // βλ. σημείωση για pending approval παρακάτω
    Severity      Severity
    SchemaVersion int         // επιτρέπει μελλοντικές αλλαγές σχήματος χωρίς breaking migration

    Payload any // ανοιχτό, ελεύθερο περιεχόμενο ανά EventType
}

type EventHandler interface {
    Handle(Event) error
}
```

**Event ID:** Πρέπει να είναι globally unique — προτιμάται **UUIDv7**
αντί για απλό UUIDv4 ή αύξοντα αριθμό, γιατί το UUIDv7 ενσωματώνει
timestamp στη δομή του (sortable by creation time) — χρήσιμο ήδη από
τώρα για σωστή σειρά replay, και απαραίτητο αργότερα για
sync/deduplication αν προστεθεί cloud sync (v0.4+).

**Γιατί typed metadata + generic payload, όχι πλήρως ανοιχτό
`map[string]any`:** Ένα εντελώς ανοιχτό schema (`Payload["cmd"]` vs
`Payload["command"]` vs `Payload["Command"]`) δημιουργεί silent typos
χωρίς compile-time έλεγχο. Κρατάμε το "σκελετό" του event (ID,
timestamp, τύπος, status, severity, schema version) typed και σταθερό,
και μόνο το ελεύθερο περιεχόμενο (`Payload`) παραμένει γενικό ανά
`EventType`.

**Σχετικά με το Status:** Το πεδίο `Status` υπάρχει ειδικά για το
Protection layer (v0.3). Όταν μια πολιτική απαιτεί έγκριση χρήστη πριν
την εκτέλεση, το event γράφεται αρχικά με `Status: PENDING_APPROVAL`.
Αν ο χρήστης απαντήσει "όχι", ενημερώνεται σε `BLOCKED`· αν απαντήσει
"ναι", ενημερώνεται σε `EXECUTED`. Χωρίς αυτό το ενδιάμεσο status, το
audit log θα καταγράφει ενέργειες ως εκτελεσμένες πριν πάρουν έγκριση —
δηλαδή θα είναι ανακριβές.

Κάθε event γράφεται πλήρως στη SQLite **πριν** σταλεί σε οποιονδήποτε
handler, ώστε μελλοντικοί καταναλωτές (π.χ. web dashboard) να μπορούν
να διαβάσουν το ιστορικό αναδρομικά, όχι μόνο live events.

## Alternatives considered

- **Plugin system από το v0.1**: απορρίφθηκε. Δικαιολογείται μόνο όταν
  υπάρχουν πολλαπλοί, άγνωστοι εκ των προτέρων συγγραφείς plugins
  (τρίτοι developers). Στο v0.1 ο μόνος "συγγραφέας" είναι ο δημιουργός
  του project — η δυναμική φόρτωση θα ήταν over-engineering για
  πρόβλημα που δεν υπάρχει ακόμα.
- **Monolithic function calls** (χωρίς κανένα interface): απορρίφθηκε
  γιατί θα δυσκόλευε την προσθήκη Protection/Intelligence layer αργότερα
  χωρίς rewrite του Recorder.

## Consequences

- Θετικό: Το v0.1 μένει απλό και γρήγορο να υλοποιηθεί.
- Θετικό: Το interface `EventHandler` παραμένει σταθερό ακόμα κι αν
  αργότερα χρειαστεί πραγματικό plugin system — η μόνη αλλαγή θα είναι
  *πώς* φορτώνονται οι handlers (hardcoded λίστα → dynamic registry),
  όχι *πώς* επικοινωνούν.
- Ρίσκο προς αποφυγή: Το `Payload` πρέπει να παραμείνει γενικό (`any`),
  όχι strongly-typed struct ανά είδος event, ώστε η προσθήκη νέων τύπων
  events να μη σπάει τον πυρήνα. Το υπόλοιπο "σκελετό" του event (ID,
  τύπος, status, severity, schema version) παραμένει typed για να
  αποφεύγονται typos σε string keys.
- Το `Status` field (EXECUTED / PENDING_APPROVAL / BLOCKED) είναι
  απαραίτητο από το v0.3 (Protection layer) ώστε το audit log να μένει
  ακριβές όταν μια ενέργεια περιμένει έγκριση χρήστη πριν εκτελεστεί.
- Μελλοντική επέκταση: Αν στο v1.0 χρειαστεί τρίτοι να γράφουν custom
  policies, ένα plugin system θα χτιστεί **πάνω** σε αυτό το μοντέλο,
  όχι αντί αυτού.
- Ανοιχτό μέλλον: Μελλοντικές εκδόσεις μπορεί να εισάγουν typed payloads
  για επιλεγμένα, υψηλής συχνότητας events, αν προκύψει πραγματική
  ανάγκη για validation. Αυτή η απόφαση δεν κλειδώνει το μέλλον προς
  καμία κατεύθυνση.