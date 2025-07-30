package constant

import "strings"

type Item string

var mineable = []Item{
	ItemIceWater,
	ItemQuartzSand,
	ItemAmmoniaIce,
	ItemSiliconCrystals,
}

var siphonable = []Item{
	ItemHydrocarbon,
	ItemLiquidNitrogen,
	ItemLiquidHydrogen,
}

const (
	ItemPreciousStones          Item = "PRECIOUS_STONES"
	ItemQuartzSand              Item = "QUARTZ_SAND"
	ItemSiliconCrystals         Item = "SILICON_CRYSTALS"
	ItemAmmoniaIce              Item = "AMMONIA_ICE"
	ItemLiquidHydrogen          Item = "LIQUID_HYDROGEN"
	ItemLiquidNitrogen          Item = "LIQUID_NITROGEN"
	ItemIceWater                Item = "ICE_WATER"
	ItemExoticMatter            Item = "EXOTIC_MATTER"
	ItemAdvancedCircuitry       Item = "ADVANCED_CIRCUITRY"
	ItemGravitonEmitters        Item = "GRAVITON_EMITTERS"
	ItemIron                    Item = "IRON"
	ItemIronOre                 Item = "IRON_ORE"
	ItemCopper                  Item = "COPPER"
	ItemCopperOre               Item = "COPPER_ORE"
	ItemAluminum                Item = "ALUMINUM"
	ItemAluminumOre             Item = "ALUMINUM_ORE"
	ItemSilver                  Item = "SILVER"
	ItemSilverOre               Item = "SILVER_ORE"
	ItemGold                    Item = "GOLD"
	ItemGoldOre                 Item = "GOLD_ORE"
	ItemPlatinum                Item = "PLATINUM"
	ItemPlatinumOre             Item = "PLATINUM_ORE"
	ItemDiamonds                Item = "DIAMONDS"
	ItemUranite                 Item = "URANITE"
	ItemUraniteOre              Item = "URANITE_ORE"
	ItemMeritium                Item = "MERITIUM"
	ItemMeritiumOre             Item = "MERITIUM_ORE"
	ItemHydrocarbon             Item = "HYDROCARBON"
	ItemAntimatter              Item = "ANTIMATTER"
	ItemFabMats                 Item = "FAB_MATS"
	ItemFertilizers             Item = "FERTILIZERS"
	ItemFabrics                 Item = "FABRICS"
	ItemFood                    Item = "FOOD"
	ItemJewelry                 Item = "JEWELRY"
	ItemMachinery               Item = "MACHINERY"
	ItemFirearms                Item = "FIREARMS"
	ItemAssaultRifles           Item = "ASSAULT_RIFLES"
	ItemMilitaryEquipment       Item = "MILITARY_EQUIPMENT"
	ItemExplosives              Item = "EXPLOSIVES"
	ItemLabInstruments          Item = "LAB_INSTRUMENTS"
	ItemAmmunition              Item = "AMMUNITION"
	ItemElectronics             Item = "ELECTRONICS"
	ItemShipPlating             Item = "SHIP_PLATING"
	ItemShipParts               Item = "SHIP_PARTS"
	ItemEquipment               Item = "EQUIPMENT"
	ItemFuel                    Item = "FUEL"
	ItemMedicine                Item = "MEDICINE"
	ItemDrugs                   Item = "DRUGS"
	ItemClothing                Item = "CLOTHING"
	ItemMicroprocessors         Item = "MICROPROCESSORS"
	ItemPlastics                Item = "PLASTICS"
	ItemPolynucleotides         Item = "POLYNUCLEOTIDES"
	ItemBiocomposites           Item = "BIOCOMPOSITES"
	ItemQuantumStabilizers      Item = "QUANTUM_STABILIZERS"
	ItemNanobots                Item = "NANOBOTS"
	ItemAiMainframes            Item = "AI_MAINFRAMES"
	ItemQuantumDrives           Item = "QUANTUM_DRIVES"
	ItemRoboticDrones           Item = "ROBOTIC_DRONES"
	ItemCyberImplants           Item = "CYBER_IMPLANTS"
	ItemGeneTherapeutics        Item = "GENE_THERAPEUTICS"
	ItemNeuralChips             Item = "NEURAL_CHIPS"
	ItemMoodRegulators          Item = "MOOD_REGULATORS"
	ItemViralAgents             Item = "VIRAL_AGENTS"
	ItemMicroFusionGenerators   Item = "MICRO_FUSION_GENERATORS"
	ItemSupergrains             Item = "SUPERGRAINS"
	ItemLaserRifles             Item = "LASER_RIFLES"
	ItemHolographics            Item = "HOLOGRAPHICS"
	ItemShipSalvage             Item = "SHIP_SALVAGE"
	ItemRelicTech               Item = "RELIC_TECH"
	ItemNovelLifeforms          Item = "NOVEL_LIFEFORMS"
	ItemBotanicalSpecimens      Item = "BOTANICAL_SPECIMENS"
	ItemCulturalArtifacts       Item = "CULTURAL_ARTIFACTS"
	ItemFrameProbe              Item = "FRAME_PROBE"
	ItemFrameDrone              Item = "FRAME_DRONE"
	ItemFrameInterceptor        Item = "FRAME_INTERCEPTOR"
	ItemFrameRacer              Item = "FRAME_RACER"
	ItemFrameFighter            Item = "FRAME_FIGHTER"
	ItemFrameFrigate            Item = "FRAME_FRIGATE"
	ItemFrameShuttle            Item = "FRAME_SHUTTLE"
	ItemFrameExplorer           Item = "FRAME_EXPLORER"
	ItemFrameMiner              Item = "FRAME_MINER"
	ItemFrameLightFreighter     Item = "FRAME_LIGHT_FREIGHTER"
	ItemFrameHeavyFreighter     Item = "FRAME_HEAVY_FREIGHTER"
	ItemFrameTransport          Item = "FRAME_TRANSPORT"
	ItemFrameDestroyer          Item = "FRAME_DESTROYER"
	ItemFrameCruiser            Item = "FRAME_CRUISER"
	ItemFrameCarrier            Item = "FRAME_CARRIER"
	ItemFrameBulkFreighter      Item = "FRAME_BULK_FREIGHTER"
	ItemReactorSolarI           Item = "REACTOR_SOLAR_I"
	ItemReactorFusionI          Item = "REACTOR_FUSION_I"
	ItemReactorFissionI         Item = "REACTOR_FISSION_I"
	ItemReactorChemicalI        Item = "REACTOR_CHEMICAL_I"
	ItemReactorAntimatterI      Item = "REACTOR_ANTIMATTER_I"
	ItemEngineImpulseDriveI     Item = "ENGINE_IMPULSE_DRIVE_I"
	ItemEngineIonDriveI         Item = "ENGINE_ION_DRIVE_I"
	ItemEngineIonDriveIi        Item = "ENGINE_ION_DRIVE_II"
	ItemEngineHyperDriveI       Item = "ENGINE_HYPER_DRIVE_I"
	ItemModuleMineralProcessorI Item = "MODULE_MINERAL_PROCESSOR_I"
	ItemModuleGasProcessorI     Item = "MODULE_GAS_PROCESSOR_I"
	ItemModuleCargoHoldI        Item = "MODULE_CARGO_HOLD_I"
	ItemModuleCargoHoldIi       Item = "MODULE_CARGO_HOLD_II"
	ItemModuleCargoHoldIii      Item = "MODULE_CARGO_HOLD_III"
	ItemModuleCrewQuartersI     Item = "MODULE_CREW_QUARTERS_I"
	ItemModuleEnvoyQuartersI    Item = "MODULE_ENVOY_QUARTERS_I"
	ItemModulePassengerCabinI   Item = "MODULE_PASSENGER_CABIN_I"
	ItemModuleMicroRefineryI    Item = "MODULE_MICRO_REFINERY_I"
	ItemModuleScienceLabI       Item = "MODULE_SCIENCE_LAB_I"
	ItemModuleJumpDriveI        Item = "MODULE_JUMP_DRIVE_I"
	ItemModuleJumpDriveIi       Item = "MODULE_JUMP_DRIVE_II"
	ItemModuleJumpDriveIii      Item = "MODULE_JUMP_DRIVE_III"
	ItemModuleWarpDriveI        Item = "MODULE_WARP_DRIVE_I"
	ItemModuleWarpDriveIi       Item = "MODULE_WARP_DRIVE_II"
	ItemModuleWarpDriveIii      Item = "MODULE_WARP_DRIVE_III"
	ItemModuleShieldGeneratorI  Item = "MODULE_SHIELD_GENERATOR_I"
	ItemModuleShieldGeneratorIi Item = "MODULE_SHIELD_GENERATOR_II"
	ItemModuleOreRefineryI      Item = "MODULE_ORE_REFINERY_I"
	ItemModuleFuelRefineryI     Item = "MODULE_FUEL_REFINERY_I"
	ItemMountGasSiphonI         Item = "MOUNT_GAS_SIPHON_I"
	ItemMountGasSiphonIi        Item = "MOUNT_GAS_SIPHON_II"
	ItemMountGasSiphonIii       Item = "MOUNT_GAS_SIPHON_III"
	ItemMountSurveyorI          Item = "MOUNT_SURVEYOR_I"
	ItemMountSurveyorIi         Item = "MOUNT_SURVEYOR_II"
	ItemMountSurveyorIii        Item = "MOUNT_SURVEYOR_III"
	ItemMountSensorArrayI       Item = "MOUNT_SENSOR_ARRAY_I"
	ItemMountSensorArrayIi      Item = "MOUNT_SENSOR_ARRAY_II"
	ItemMountSensorArrayIii     Item = "MOUNT_SENSOR_ARRAY_III"
	ItemMountMiningLaserI       Item = "MOUNT_MINING_LASER_I"
	ItemMountMiningLaserIi      Item = "MOUNT_MINING_LASER_II"
	ItemMountMiningLaserIii     Item = "MOUNT_MINING_LASER_III"
	ItemMountLaserCannonI       Item = "MOUNT_LASER_CANNON_I"
	ItemMountMissileLauncherI   Item = "MOUNT_MISSILE_LAUNCHER_I"
	ItemMountTurretI            Item = "MOUNT_TURRET_I"
	ItemShipProbe               Item = "SHIP_PROBE"
	ItemShipMiningDrone         Item = "SHIP_MINING_DRONE"
	ItemShipSiphonDrone         Item = "SHIP_SIPHON_DRONE"
	ItemShipInterceptor         Item = "SHIP_INTERCEPTOR"
	ItemShipLightHauler         Item = "SHIP_LIGHT_HAULER"
	ItemShipCommandFrigate      Item = "SHIP_COMMAND_FRIGATE"
	ItemShipExplorer            Item = "SHIP_EXPLORER"
	ItemShipHeavyFreighter      Item = "SHIP_HEAVY_FREIGHTER"
	ItemShipLightShuttle        Item = "SHIP_LIGHT_SHUTTLE"
	ItemShipOreHound            Item = "SHIP_ORE_HOUND"
	ItemShipRefiningFreighter   Item = "SHIP_REFINING_FREIGHTER"
	ItemShipSurveyor            Item = "SHIP_SURVEYOR"
	ItemShipBulkFreighter       Item = "SHIP_BULK_FREIGHTER"
)

func (i Item) IsOre() bool {
	return strings.HasSuffix(string(i), "_ORE")
}

func (i Item) isInList(list []Item) bool {
	for _, item := range list {
		if item == i {
			return true
		}
	}
	return false
}

func (i Item) IsMineable() bool {
	return i.IsOre() || i.isInList(mineable)
}

func (i Item) IsRefinable() bool {
	return i.IsOre()
}

func (i Item) IsSiphonable() bool {
	return i.isInList(siphonable)
}
