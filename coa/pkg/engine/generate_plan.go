package engine

import (
	"coa/pkg/pilot" // Importiamo i tipi definiti nel pilota
	"encoding/json"
	"os"
)

func GeneratePlan(yamlSteps []pilot.YamlStep, familyID string, isRemaster bool, workPath string) error {
	var plan OAPlan

	// Definiamo l'utente classico "live/live"
	// In futuro questo verrà da pilot.LoadConfig()
	defaultUser := pilot.User{
		Login:    "live",
		Password: "$6$rounds=4096$salt$UQ7.P... (hash di 'live')",
		Home:     "/home/live",
		Shell:    "/bin/bash",
		Groups:   []string{"sudo", "audio", "video", "cdrom", "plugdev", "netdev"},
		UID:      1000,
		GID:      1000,
	}

	for _, step := range yamlSteps {
		switch step.Command {

		case "oa_mount_logic":
			// Esplode la vecchia logica del C in tanti task JSON
			plan.Plan = append(plan.Plan, expandMountLogic(workPath)...)

		case "oa_users":
			plan.Plan = append(plan.Plan, OATask{
				Command:    "oa_shell",
				Info:       "Creazione home directory da /etc/skel",
				RunCommand: "mkdir -p " + workPath + "/liveroot/home/live && cp -a " + workPath + "/liveroot/etc/skel/. " + workPath + "/liveroot/home/live/",
			})

			plan.Plan = append(plan.Plan, OATask{
				Command:    "oa_users",
				Info:       "Iniezione identità live/live",
				PathLiveFs: workPath,
				Users:      []pilot.User{defaultUser},
			})

		case "oa_umount":
			plan.Plan = append(plan.Plan, OATask{
				Command:    "oa_umount",
				Info:       "Pulizia finale dei mount",
				PathLiveFs: workPath,
			})

		default:
			// Passaggio diretto dallo YAML al JSON (permette a debian.yaml di chiamare qualsiasi verbo C)
			plan.Plan = append(plan.Plan, OATask{
				Command:    step.Command, // <--- Usa il comando originale dello YAML!
				Info:       step.Description,
				RunCommand: step.RunCommand,
				Chroot:     step.Chroot,
				PathLiveFs: workPath,
				Path:       step.Path, // <--- Passa il parametro
				Src:        step.Src,  // <--- Passa il parametro
				Dst:        step.Dst,  // <--- Passa il parametro
			})
		}
	}

	// Scrittura del file JSON finale
	return savePlan(plan)
}

func savePlan(plan OAPlan) error {
	os.MkdirAll("/tmp/coa", 0755)
	file, _ := json.MarshalIndent(plan, "", "  ")
	return os.WriteFile("oa-plan.json", file, 0644)
}
