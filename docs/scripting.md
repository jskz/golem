# JavaScript API

Golem embeds a JavaScript engine and extends to it an API for registering event handlers and otherwise influencing clients and other gameplay objects.

## Globals

The **Golem** global provides access to the following properties:

| Type | Name | Arguments | Description | Example 
| --- | --- | --- | ----------- | --- | 
| Method | broadcast | `message`: **String** | Sends `message` to all connected and in-game players, without a filter. | ```Golem.broadcast("The sky is falling; the server is shutting down!\r\n");```
| Method | registerPlayerCommand | `command`: **String**, `callback`: function(`ch`: **Character**, `args`: **String**) | Registers a player interpreter command `command` if a system default does not exist.  If a scripted `command` already exists, its callback is overriden.  The callback is executed with the calling player character handle and any command arguments unsplit. | `Golem.registerPlayerCommand('echo', function(ch, args) { ch.send("Your arguments: " + args + "\r\n"); });`
| Method | registerSpellHandler | `spell`: **String**, `callback`: function(`ch`: **Character**, `args`: **String**) | Registers or overwrites the callback handler for a specific spell, if that spell is defined.  *This API will be subject to major change.* | `Golem.registerSpellHandler('cure light', function(ch, args) { Golem.game.damage(null, ch, false, -(~~(Math.random() * 5) + 5), Golem.Combat.DamageTypeExotic); ch.send("{WYou feel a little bit better.{x\r\n"); });`
| Field | game: **Game** |  | Provides access to many global gameplay session values and utility methods.   Refer Game section. | `Golem.game.fights.head.value.participants` 

## Game

A convenient reference to the `game` singleton is exposed through the `Golem.game` field with the following properties:


| Type |  Name | Arguments | Description | Example
| --- | --- | --- | --- | --
| Method | damage | *`origin`?: **Character*** = **null**, `target`: **Character**, `display`: **Boolean**, `amount`: **Integer**, `damageType`: **GolemDamageType** | Inflicts `amount` damage of type `damageType` on `target`.  If `display` is true and `origin` is also a `Character`, then this damage is broadcast to the appropriate players as combat output. 
| Field | fights: **LinkedList\<Combat\>** | | All active combat sessions in the game. | 
| Field | characters: **LinkedList\<Character\>** | | All active character instances, PC or NPC, in the game.

## Character

PCs and NPCs both share a common Character data type with the following properties exposed to the scripting API:

| Type |  Name | Arguments | Description | Example
| --- | --- | --- | --- | ---
| Method | send | `message`: **String** | Sends `message` exclusively to this character instance. | ```ch.send("Hello world!\r\n");```
| Method | findCharacterInRoom: **Character**? |  `name`: **String** | Tries to find a character by name in the same room as this character, may return **null**. | `const target = ch.findCharacterInRoom('monster');`

## Room

| Type |  Name | Arguments | Description | Example
| --- | --- | --- | --- | ---
| Method | broadcast | `message`: **String**, *`filter`*?: function(`ch`: **Character**) = **null** | Sends `message` to characters in room for which `filter(ch) === true` or all characters if filter is **null**. | ```ch.Room.broadcast("Message to other people in this room", rch => !rch.isEqual(ch));```

## Combat

| Type |  Name | Arguments | Description
| --- | --- | --- | --- |
| Field | participants: [**Character**] | | An array of characters participating in a combat instance. |


## LinkedList

Iterating an API-exposed linked list with a `Combat` type:

```js
// Damage everybody who is currently in any fight for 50 damage of "exotic" type
for(let iter = Golem.game.fights.head; iter != null; iter = iter.next) {
    const fight = iter.value;

    // fight.participants is an [Character] Array of all involved characters
    for(let f = 0; f < fight.participants.length; f++) {
        fight.participants[f].send("A developer poked you in the eye!\r\n");
        Golem.game.damage(null, fight.participants[f], false, 50, Golem.Combat.DamageTypeExotic);
    }
}

```