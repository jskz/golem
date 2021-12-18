# 1.0 Roadmap

This will be a working document to track major feature development goals for each milestone release towards a 1.0 of a scope manageable with clearly defined chunks in my limited spare time.

## 0.1 Milestones

- [x] Simplify the system versus local configuration paths
- [x] README: Refactor requirements and setup sections as multi-section and qualify "Docker-based Setup" subsection
- [x] Serialization for persistent lists of object instances: upserting object instances and syncing with detach the appropriate relation
- [x] If output overflows, instead of gameplay prompt, page to a reasonable (configurable, default telnet window size?) output limit and append a next/previous and cursor read % prompt
- [x] Create a `Dungeon` data type for instancing multi-level mazes with metadata and reasonably "difficult" start/end goals representing passages to the next maze floor
- [x] Gameplay: implement completely an initially available job-unique skill for job, and another job-unique skill available after the 5th level (8 specials total)

## 0.2 Milestones

- [x] Host a persistent pre-alpha instance and welcome limited testing: `mud.jskarzin.org:4000`
- [x] Exit flags: closing and opening doors, movement obstruction
- [x] Clean up the inconsistent title case method and field naming stemming from the scripting field mapper change
- [x] Extend to scripting and entities a timer-based "effects" model for behaviours like buffs/enchantments on equipment or spellcasting, permanent or with expiration
- [x] Implement a few NPC flags for behaviours: stay in an area, move, aggressive/hostile towards PCs, etc.
- [x] Groups: forming and unforming groups with other players (and NPCs), interaction commands/combat mechanics/experience splitting
- [x] Serialization for player proficiencies with flexibility for multi-job system in future

## 0.3 Scripting and Shops, Plane Development Milestones

- [x] Add a mechanism for NPCs (with flexibility for PCs in future?) to operate shops that players can buy from, sell?
- [x] Webhook facility for scripting API
- [x] Create a new model `plane` representing a 2D array by dimensions and a blob of integer terrain pivots.  The data structure for a plane may evolve towards a quadtree but until post-1.0 will remain simple
- [x] Create a new model `terrain` to store information about unique terrain types: name, a glyph used to represent that terrain type in a 2D map, movement cost to traverse a unit of this terrain type, etc.
- [x] Allow players to navigate a plane-based virtual room, like maze-based virtual rooms, with an ANSI art representation of the nearby landscape
- [x] Allow for a mechanism by which each app startup will generate (preset or random?) varied `Dungeon` instances and then create virtual exits between chosen planes

## 0.4 Pivot for Scope, Simplify

- [x] Gameplay pivots: harsher penalties always drop all loot, gold on death
- [x] Ensure that the caster's proficiency details are passed into the script handler context for a given skill or spell for logic based on proficiency %
- [x] Lean into content generation for the overworld map: Perlin noise to simulate some islands with a terrain gradient between shores, field, and trees of various density
- [ ] Complete and merge the quadtree branch, "district" metadata for things like regional terrain name overrides, QueryRect for players/other objects, have "harder" regions with aggressive or passive NPCs, auto-generated rewards; no "resets" on overworld, run planar reset scripts on an interval instead

## 0.5 Online Creation Milestones

- [ ] Online Creation (OLC) command set: redit, medit, oedit, reset, digging two-way exits from a room
- [ ] World persistence updates: currently mobile, room, and object instances have only read operations defined

## 0.6 Procedural Content Development Milestones

- [ ] Library of templated parts to make generated maze-based dungeons more interesting without breaking "solvability" of the mazes
- [ ] Develop sufficient base data for procedural loot tables for mazes and overworld

## 0.7 Gameplay General Content Development Milestones

- [x] Stat rolling at character creation, stat buffing/nerfs on effect/enchantments
- [ ] Extra commands and UX for gameplay: combat experience, leveling/healing/item consumption gameplay loop considerations: "is it fun to hack-and-slash and navigate this as a game?"
- [ ] Gameplay and content development: ensure that each player job choice has at least ten unique skills and/or spells

## 0.8 The (Problematic) Human Element Development Milestones

- [ ] Socials: flavour text commands for socializing in-room like grin, nod, laugh
- [ ] Enforcement: bans on username and host (IP? allow covering prefix with single ban?)

## 0.9 Tying It All Together Milestones

- [ ] Portals?  Planes and zones shouldn't need "soft-coded" scripts for mutual exits, portals which cover all plane/zone cases
- [ ] "Town square" job/message board which both UX resembles a popular old "boards" plugin, and also provides a generated quest gameplay mechanic: venture into level/difficulty-based random choice of floor X of generated dungeon Y to do some combination of obtain/kill/charm a chosen item/creature in order to return for Z quest points

## 1.0 Release Milestones

- [ ] Harden existing telnet implementation and begin fuzzing; we'll rewrite the telnet implementation as a post-1.0 goal
- [ ] TBD; errata, polish, suggestions uncovered en route
