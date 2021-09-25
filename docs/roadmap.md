# 1.0 Roadmap

This will be a working document to track major feature development goals for each milestone release towards a 1.0 of a scope manageable with clearly defined chunks in my limited spare time.

## 0.1 Milestones

- [ ] Simplify the system versus local configuration paths
- [x] README: Refactor requirements and setup sections as multi-section and qualify "Docker-based Setup" subsection
- [x] [Serialization for persistent lists of object instances: upserting object instances and syncing with detach the appropriate relation](https://github.com/jskz/golem/pull/2)
- [ ] If output overflows, instead of gameplay prompt, page to a reasonable (configurable, default telnet window size?) output limit and append a next/previous and cursor read % prompt
- [ ] Create a `Dungeon` data type for instancing multi-level mazes with metadata and reasonably "difficult" start/end goals representing passages to the next maze floor
- [ ] Gameplay: implement completely an initially available job-unique skill for job, and another job-unique skill available after the 5th level (8 specials total)

## 0.2 Milestones

- [x] Exit flags: closing and opening doors, movement obstruction
- [ ] Clean up the inconsistent title case method and field naming stemming from the scripting field mapper change
- [ ] Extend to scripting and entities a timer-based "effects" model for behaviours like buffs/enchantments on equipment or spellcasting, permanent or with expiration
- [ ] Implement a few NPC flags for behaviours: stay in an area, move, aggressive/hostile towards PCs, etc.

## 0.3 Telnet Done Right Milestones

- [ ] Telnet IAC parsing/session handling implementation overhaul

## 0.4 Telnet Zlib Compression Milestones

- [ ] [MCCP2 or MCCP3 support](https://mudhalla.net/tintin/protocols/mccp/)
- [ ] Find a solution to perform some fuzz testing of the telnet implementation
- [ ] README subsection for traditional compile and run with existing services

## 0.5 Online Creation Milestones

- [ ] Online Creation (OLC) command set: redit, medit, oedit, reset, digging two-way exits from a room
- [ ] World persistence updates: currently mobile, room, and object instances have only read operations defined

## 0.6 Plane Development Milestones

- [ ] Create a new model `plane` representing a 2D array by dimensions and a blob of integer terrain pivots.  The data structure for a plane may evolve towards a quadtree but until post-1.0 will remain simple
- [ ] Create a new model `terrain` to store information about unique terrain types: name, a glyph used to represent that terrain type in a 2D map, movement cost to traverse a unit of this terrain type, etc.
- [ ] Allow players to navigate a plane-based virtual room, like maze-based virtual rooms, with an ANSI art representation of the nearby landscape
- [ ] Allow for a mechanism by which each app startup will generate (preset or random?) varied `Dungeon` instances and then create virtual exits between chosen planes

## 0.7 Plane-based Content Development Milestones

- [ ] Procedural object instance generation for "ephemeral" items that do not have prototype objects, only instances; string tables will suffice for 1.0
- [ ] Procedural NPC monster instance generation for similar creature Character instances to populate generated maze zones

## 0.8 Gameplay Content Development

- [ ] Gameplay development and balancing: experience curves and enchantment tweaks for races and jobs
- [ ] Extra commands and UX for gameplay: combat experience, leveling/healing/item consumption gameplay loop considerations: "is it fun to hack-and-slash and navigate this as a game?"
- [ ] Gameplay and content development: Ensure that each player job choice has at least ten unique skills and/or spells available by the hero level (50)
- [ ] Socials: flavour text commands for socializing in-room like grin, nod, laugh

## 0.9 Tying It All Together Milestones

- [ ] Allow for a repeatable OLC-available mechanism to declare a crafted zone to have a virtual exit to some plane: fixed position, random range, feature-based?
- [ ] Content: "town square" board which both UX resembles a popular old "boards" plugin, and also provides a generated quest gameplay mechanic: venture into level/difficulty-based random choice of floor X of generated dungeon Y to do some combination of obtain/kill/charm a chosen item/creature in order to return for Z quest points

## 1.0 Release Milestones

- [ ] Open instance of the server to alpha testing and tag the most recent master commit as 1.0 when a stranger first joins the game
