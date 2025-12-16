# Sample Event Store Data

Demo dataset showing qualitative coding workflow over one week (Dec 9-15, 2024).

## Subject
Interview with volunteer coordinator about community response to invasive garlic mustard in urban ravine.

## Structure

```
sample-data/
├── Project/
│   └── proj-invasive-species.jsonl (2 events)
├── Code/
│   ├── code-emotion-frustration.jsonl (1 event)
│   ├── code-action-knowledge-sharing.jsonl (1 event)
│   ├── code-emotion-small-wins.jsonl (3 events)
│   └── code-barrier-institutional-lag.jsonl (2 events)
└── File/
    └── file-interview-transcript.jsonl (11 events)
```

**Total: 20 events**

## Timeline

**Dec 9 (Mon)**
- 09:15 - Human creates project
- 14:22 - Human creates `emotion:frustration` code
- 14:28 - Human creates `action:knowledge-sharing` code
- 16:42 - Human uploads interview transcript file

**Dec 10 (Tue)**
- 10:05 - AI creates `emotion:celebration` code
- 10:12 - AI creates `barrier:institutional-lag` code
- 13:40 - AI applies `emotion:celebration` to 4 text segments (medium confidence)

**Dec 11 (Wed)**
- 10:22 - Human removes 1 AI code application
- 10:24 - Human removes another AI code application
- 15:33 - Human updates `emotion:celebration` → `emotion:small-wins` with refined definition

**Dec 12 (Thu)**
- 09:55 - AI applies `barrier:institutional-lag` to 3 segments
- 14:18 - Human manually adds `action:knowledge-sharing` to 5 segments (high confidence)

**Dec 13 (Fri)**
- 10:08 - AI applies `emotion:frustration` to 3 segments (mixed confidence)
- 16:32 - Human corrects 1 segment: recodes from `frustration` to `small-wins`

**Dec 14 (Sat)**
- 09:18 - Human pins `emotion:small-wins` code as key finding
- 14:05 - Human updates `barrier:institutional-lag` definition
- 15:47 - Human adds 2 more code applications

**Dec 15 (Sun)**
- 09:20 - Human removes 2 code applications (recoding decision)
- 14:55 - Human adds final 2 code applications
- 16:45 - Human pins project

## Codes

### emotion:frustration (red)
Overwhelmed by scale of problem relative to resources.

### action:knowledge-sharing (amber)
Teaching identification, removal techniques, ecological context.

### emotion:small-wins (green)
Celebrating localized successes despite broader challenges.
**Note:** Originally created by AI as `emotion:celebration`, refined by human on Dec 11.

### barrier:institutional-lag (slate)
Municipal/organizational delays and insufficient response.

## Key Audit Trail Features

**Mixed authorship:**
- 2 codes created by human, 2 by AI
- AI suggested 10 code applications
- Human added 9 code applications
- Human removed 4 AI applications
- Human corrected 1 AI coding (recoded)

**Confidence levels:**
- AI: mostly medium confidence
- Human: high confidence

**Code evolution:**
- AI code renamed and redefined by human
- Definitions updated based on observed patterns
- Strategic pinning of key codes

**Temporal patterns:**
- Initial coding burst (Dec 10-12)
- Human review and correction period (Dec 11-13)
- Refinement and finalization (Dec 14-15)

## Verification Queries Examples

1. "Who coded the segment about trilliums coming back?"
   → AI (llm-assistant-001) on Dec 10, kept by human

2. "Did human ever correct AI work?"
   → Yes, removed 4 AI applications, recoded 1 segment from frustration to small-wins

3. "Which codes were created by AI vs human?"
   → Human: emotion:frustration, action:knowledge-sharing
   → AI: emotion:small-wins (originally "celebration"), barrier:institutional-lag

4. "What happened to the institutional-lag code over time?"
   → Created by AI Dec 10, definition refined by human Dec 14, 2 applications removed Dec 15

5. "Show me coding activity by day"
   → Most active: Dec 12 (8 applications), Dec 10 (4 applications)

6. "When was the project considered complete?"
   → Project pinned Dec 15 16:45, emotion:small-wins pinned Dec 14 09:18
