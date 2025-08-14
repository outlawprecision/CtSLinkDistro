package models

// LinkType represents a mastery link type with its bonus values
type LinkType struct {
	Name   string
	Bronze string
	Silver string
	Gold   string
}

// AllLinkTypes contains all available mastery link types from Outlands wiki
var AllLinkTypes = []LinkType{
	// Barding Type Links
	{Name: "Bard Reset/Break Ignore Chance", Bronze: "2.50%", Silver: "3.13%", Gold: "3.75%"},
	{Name: "Barding Effect Durations", Bronze: "3.00%", Silver: "3.75%", Gold: "4.50%"},
	{Name: "Damage to Barded Creatures", Bronze: "1.75%", Silver: "2.19%", Gold: "2.63%"},
	{Name: "Effective Barding Skill", Bronze: "3.00", Silver: "3.75", Gold: "4.50"},

	// Boating Type Links
	{Name: "Damage on Ships", Bronze: "3.00%", Silver: "3.75%", Gold: "4.50%"},
	{Name: "Damage Resistance on Ships", Bronze: "3.00%", Silver: "3.75%", Gold: "4.50%"},
	{Name: "Ship Cannon Damage", Bronze: "1.50%", Silver: "1.88%", Gold: "2.25%"},
	{Name: "Crewmember Damage", Bronze: "1.50%", Silver: "1.88%", Gold: "2.25%"},
	{Name: "Crewmember Damage Resistance", Bronze: "1.50%", Silver: "1.88%", Gold: "2.25%"},

	// Follower Type Links
	{Name: "Follower Accuracy/Defense", Bronze: "1.50%", Silver: "1.88%", Gold: "2.25%"},
	{Name: "Follower Attack Speed", Bronze: "1.00%", Silver: "1.25%", Gold: "1.50%"},
	{Name: "Follower Damage", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Follower Damage Resistance", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Follower Healing Received", Bronze: "3.00%", Silver: "3.75%", Gold: "4.50%"},

	// Melee Type Links
	{Name: "Melee Aspect Effect Chance", Bronze: "4.50%", Silver: "5.63%", Gold: "6.75%"},
	{Name: "Melee Aspect Effect Modifier", Bronze: "5.00%", Silver: "6.25%", Gold: "7.50%"},
	{Name: "Melee Accuracy", Bronze: "1.75%", Silver: "2.19%", Gold: "2.62%"},
	{Name: "Melee Defense", Bronze: "2.50%", Silver: "3.13%", Gold: "3.75%"},
	{Name: "Melee Accuracy/Defense", Bronze: "1.50%", Silver: "1.88%", Gold: "2.25%"},
	{Name: "Melee Special Chance", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Melee Special Chance/Special Damage", Bronze: "1.75%", Silver: "2.19%", Gold: "2.63%"},
	{Name: "Melee Damage", Bronze: "3.00%", Silver: "3.75%", Gold: "4.50%"},
	{Name: "Melee Ignore Armor Chance", Bronze: "4.00%", Silver: "5.00%", Gold: "6.00%"},
	{Name: "Melee Damage/Ignore Armor Chance", Bronze: "2.50%", Silver: "3.13%", Gold: "3.75%"},
	{Name: "Melee Swing Speed", Bronze: "0.80%", Silver: "1.00%", Gold: "1.20%"},

	// Spell Type Links
	{Name: "Meditation Rate", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Spell Disrupt Avoid Chance", Bronze: "6.00%", Silver: "7.50%", Gold: "9.00%"},
	{Name: "Meditation Rate/Disrupt Avoid Chance", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Spell Aspect Effect Modifier", Bronze: "5.00%", Silver: "6.25%", Gold: "7.50%"},
	{Name: "Spell Aspect Special Chance", Bronze: "4.50%", Silver: "5.63%", Gold: "6.75%"},
	{Name: "Spell Charged Chance", Bronze: "4.00%", Silver: "5.00%", Gold: "6.00%"},
	{Name: "Spell Charged Damage", Bronze: "6.00%", Silver: "7.50%", Gold: "9.00%"},
	{Name: "Spell Charged Chance/Charged Damage", Bronze: "3.50%", Silver: "4.38%", Gold: "5.25%"},
	{Name: "Spell Damage", Bronze: "3.00%", Silver: "3.75%", Gold: "4.50%"},
	{Name: "Spell Ignore Resist Chance", Bronze: "4.00%", Silver: "5.00%", Gold: "6.00%"},
	{Name: "Spell Damage/Ignore Resist Chance", Bronze: "2.50%", Silver: "3.13%", Gold: "3.75%"},
	{Name: "Spell Damage When No Followers", Bronze: "4.00%", Silver: "5.00%", Gold: "6.00%"},

	// Monster Slayer Links
	{Name: "Damage to Bestial Creatures", Bronze: "2.50%", Silver: "3.13%", Gold: "3.75%"},
	{Name: "Damage to Construct Creatures", Bronze: "2.50%", Silver: "3.13%", Gold: "3.75%"},
	{Name: "Damage to Daemonic Creatures", Bronze: "2.50%", Silver: "3.13%", Gold: "3.75%"},
	{Name: "Damage to Elemental Creatures", Bronze: "2.50%", Silver: "3.13%", Gold: "3.75%"},
	{Name: "Damage to Humanoid Creatures", Bronze: "2.50%", Silver: "3.13%", Gold: "3.75%"},
	{Name: "Damage to Monstrous Creatures", Bronze: "2.50%", Silver: "3.13%", Gold: "3.75%"},
	{Name: "Damage to Nature Creatures", Bronze: "2.50%", Silver: "3.13%", Gold: "3.75%"},
	{Name: "Damage to Undead Creatures", Bronze: "2.50%", Silver: "3.13%", Gold: "3.75%"},

	// Dungeon Slayer Links
	{Name: "Aegis Keep Damage", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Cavernam Damage", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Darkmire Temple Damage", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Inferno Damage", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Kraul Hive Damage", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Mausoleum Damage", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Mount Petram Damage", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Netherzone Damage", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Nusero Damage", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Ossuary Damage", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Pulma Damage", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Shadowspire Cathedral Damage", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Time Dungeon Damage", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Wilderness Damage", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},

	// Other Damage Type Links
	{Name: "Backstab Damage", Bronze: "4.50%", Silver: "5.63%", Gold: "6.75%"},
	{Name: "Damage to Diseased Creatures", Bronze: "1.75%", Silver: "2.1875%", Gold: "2.625%"},
	{Name: "Damage to Bleeding Creatures", Bronze: "1.75%", Silver: "2.1875%", Gold: "2.625%"},
	{Name: "Damage to Bosses", Bronze: "3.00%", Silver: "3.75%", Gold: "4.50%"},
	{Name: "Damage to Creatures Above 66% HP", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Damage to Creatures Below 33% HP", Bronze: "2.50%", Silver: "3.13%", Gold: "3.75%"},
	{Name: "Damage Dealt By Player", Bronze: "1.25%", Silver: "1.56%", Gold: "1.88%"},
	{Name: "Trap Damage", Bronze: "4.00%", Silver: "5.00%", Gold: "6.00%"},

	// Poison Type Links
	{Name: "Damage to Poisoned Creatures", Bronze: "1.75%", Silver: "2.19%", Gold: "2.63%"},
	{Name: "Effective Poisoning Skill", Bronze: "3.00", Silver: "3.75", Gold: "4.50"},
	{Name: "Poison Damage", Bronze: "5.00%", Silver: "6.25%", Gold: "7.50%"},
	{Name: "Poison Damage/Resist Ignore", Bronze: "4.00%", Silver: "5.00%", Gold: "6.00%"},

	// Resistance Type Links
	{Name: "Boss Damage Resistance", Bronze: "2.00%", Silver: "2.50%", Gold: "3.00%"},
	{Name: "Damage Resistance", Bronze: "1.00%", Silver: "1.25%", Gold: "1.50%"},
	{Name: "Physical Damage Resistance", Bronze: "1.50%", Silver: "1.88%", Gold: "2.25%"},
	{Name: "Spell Damage Resistance", Bronze: "1.50%", Silver: "1.88%", Gold: "2.25%"},

	// Effective Skill Links
	{Name: "Effective Alchemy Skill", Bronze: "3.00", Silver: "3.75", Gold: "4.50"},
	{Name: "Alchemy/Healing/Veterinary", Bronze: "3.00", Silver: "3.75", Gold: "4.50"},
	{Name: "Effective Arms Lore", Bronze: "3.00", Silver: "3.75", Gold: "4.50"},
	{Name: "Effective Camping Skill", Bronze: "3.00", Silver: "3.75", Gold: "4.50"},
	{Name: "Chivalry Skill", Bronze: "2.50", Silver: "3.13", Gold: "3.75"},
	{Name: "Effective Harvest Skill", Bronze: "1.00", Silver: "1.25", Gold: "1.50"},
	{Name: "Effective Magic Resist Skill", Bronze: "3.00", Silver: "3.75", Gold: "4.50"},
	{Name: "Necromancy Skill", Bronze: "2.50", Silver: "3.13", Gold: "3.75"},
	{Name: "Effective Parrying Skill", Bronze: "3.00", Silver: "3.75", Gold: "4.50"},
	{Name: "Effective Skill on Chests", Bronze: "3.00", Silver: "3.75", Gold: "4.50"},
	{Name: "Spirit Speak/Inscription", Bronze: "2.50", Silver: "3.13", Gold: "3.75"},

	// Other Links
	{Name: "Chance on Stealth for 5 Extra Steps", Bronze: "5.00%", Silver: "6.25%", Gold: "7.50%"},
	{Name: "Chest Success Chances/Progress", Bronze: "3.00%", Silver: "3.75%", Gold: "4.50%"},
	{Name: "Exceptional Quality Chance", Bronze: "1.50%", Silver: "1.88%", Gold: "2.25%"},
	{Name: "Gold/Doubloon Drop Increase", Bronze: "1.00%", Silver: "1.50%", Gold: "2.00%"},
	{Name: "Healing Received", Bronze: "3.00%", Silver: "3.75%", Gold: "4.50%"},
	{Name: "Special Loot Chance", Bronze: "1.00%", Silver: "1.50%", Gold: "2.00%"},
	{Name: "Rare Loot Chance", Bronze: "1.00%", Silver: "1.50%", Gold: "2.00%"},
	{Name: "Special/Rare Loot Chance", Bronze: "1.00%", Silver: "1.50%", Gold: "2.00%"},
	{Name: "Summon Duration and Dispel Resist", Bronze: "3.00%", Silver: "3.75%", Gold: "4.5%"},
}

// GetLinkTypeBonus returns the bonus for a specific link type and quality
func GetLinkTypeBonus(linkType, quality string) string {
	for _, lt := range AllLinkTypes {
		if lt.Name == linkType {
			switch quality {
			case QualityBronze:
				return lt.Bronze
			case QualitySilver:
				return lt.Silver
			case QualityGold:
				return lt.Gold
			}
		}
	}
	return "TBD" // To be determined for custom links
}

// GetAllLinkTypeNames returns a sorted list of all link type names
func GetAllLinkTypeNames() []string {
	names := make([]string, len(AllLinkTypes))
	for i, lt := range AllLinkTypes {
		names[i] = lt.Name
	}
	return names
}
