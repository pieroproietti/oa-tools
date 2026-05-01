package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"coa/pkg/pilot"
	"coa/pkg/utils"
)

// GeneratePlan converte lo YAML in JSON. Ora accetta anche stopAfter!
func GeneratePlan(yamlSteps []pilot.YamlStep, familyID string, isRemaster bool, workPath string, finalIsoPath string, stopAfter string) (string, error) {
	var plan OAPlan

	defaultUser := pilot.User{
		Login:    "live",
		Password: "$6$oa-tools$uTKAYeAVn.Y.Dy2To6HXsHt1Gt4HpMghmOV93a46jFY7hkAQ3tk7eRTKjcvSYDf5sOf3qnKzyyPYXurKp9ST3.",
		Home:     "/home/live",
		Shell:    "/bin/bash",
		Groups:   []string{"sudo", "audio", "video", "cdrom", "plugdev", "netdev"},
		UID:      1000,
		GID:      1000,
	}

	var hitBreakpoint bool

	for _, step := range yamlSteps {

		if hitBreakpoint && step.Name != "coa-cleanup" {
			continue
		}

		// --- IL PONTE: Sostituzione dinamica del percorso ISO ---
		currentRunCommand := step.RunCommand
		if strings.Contains(currentRunCommand, "${ISO_OUTPUT}") {
			currentRunCommand = strings.ReplaceAll(currentRunCommand, "${ISO_OUTPUT}", finalIsoPath)
		}

		// --- INFO DINAMICA: Mostriamo il nome reale nel log ---
		currentDescription := step.Description
		if strings.Contains(currentDescription, "${ISO_NAME}") {
			currentDescription = strings.ReplaceAll(currentDescription, "${ISO_NAME}", filepath.Base(finalIsoPath))
		}

		switch step.Command {

		case "oa_mount_logic":
			plan.Plan = append(plan.Plan, expandMountLogic(workPath)...)

		case "oa_users":
			plan.Plan = append(plan.Plan, OATask{
				Command:    "oa_shell",
				Info:       "Creazione home directory da /etc/skel",
				RunCommand: "mkdir -p " + workPath + "/liveroot/home/live && cp -a " + workPath + "/liveroot/etc/skel/. " + workPath + "/liveroot/home/live/",
			})

			// Recuperiamo i gruppi "specchiati" dall'host (es. audio, video, docker, wheel)
			mirroredGroups := utils.GetUserGroups()

			// Iniettiamo i gruppi nell'utente di default prima di passarlo al piano
			// In questo modo oa_users riceverà l'array JSON corretto
			defaultUser.Groups = mirroredGroups

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
			plan.Plan = append(plan.Plan, OATask{
				Command:    step.Command,
				Info:       currentDescription, // Usiamo la descrizione risolta
				RunCommand: currentRunCommand,  // IMPORTANTE: Usiamo il comando risolto!
				Chroot:     step.Chroot,
				PathLiveFs: workPath,
				Path:       step.Path,
				Src:        step.Src,
				Dst:        step.Dst,
			})
		}

		if stopAfter != "" && step.Name == stopAfter {
			fmt.Printf("\n\033[1;33m[ENGINE] 🛑 Breakpoint '%s' elaborato. Generazione JSON accorciata.\033[0m\n", step.Name)
			hitBreakpoint = true
		}
	}

	return savePlan(plan)
}

func savePlan(plan OAPlan) (string, error) {
	targetDir := "/tmp/coa"
	targetFile := "oa-plan.json"
	fullPath := filepath.Join(targetDir, targetFile)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}

	file, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(fullPath, file, 0644); err != nil {
		return "", err
	}

	return fullPath, nil
}
