package engine

import (
	"os"
	"path/filepath"
)

// expandMountLogic trasforma la vecchia logica statica del C in una sequenza di task JSON dinamici.
func expandMountLogic(basePath string) []OATask {
	var tasks []OATask
	liveroot := filepath.Join(basePath, "liveroot")
	overlay := filepath.Join(basePath, ".overlay")

	// 1. SETUP STRUTTURA: Creiamo le cartelle di base per l'ambiente di rimasterizzazione
	baseDirs := []string{
		liveroot,
		overlay,
		filepath.Join(overlay, "upperdir"),
		filepath.Join(overlay, "workdir"),
		filepath.Join(overlay, "lowerdir"),
	}
	for _, d := range baseDirs {
		tasks = append(tasks, OATask{Command: "oa_mkdir", Path: d, Info: "Setup base path"})
	}

	// 2. COPIE FISICHE: Necessarie per rendere il chroot funzionale e bootabile
	// Copiamo /etc (configurazioni) e /boot (kernel/initrd)
	tasks = append(tasks, OATask{Command: "oa_cp", Src: "/etc", Dst: liveroot, Info: "Copia fisica /etc"})
	tasks = append(tasks, OATask{Command: "oa_cp", Src: "/boot", Dst: liveroot, Info: "Copia fisica /boot"})

	// Copia dei symlink del kernel dalla root dell'host al liveroot
	// Usiamo os.Lstat per controllare l'esistenza dei symlink senza seguirli
	for _, link := range []string{"vmlinuz", "initrd.img", "vmlinuz.old", "initrd.img.old"} {
		src := "/" + link
		if _, err := os.Lstat(src); err == nil {
			tasks = append(tasks, OATask{
				Command: "oa_cp",
				Src:     src,
				Dst:     filepath.Join(liveroot, link),
				Info:    "Copia symlink: " + link,
			})
		}
	}

	// 3. BIND MOUNTS DINAMICI: Proiettiamo il sistema host nel chroot (Read-Only)
	// Controlliamo cosa esiste davvero sull'host per essere arch-agnostic (RISC-V, x86, etc)
	entries := []string{"bin", "sbin", "lib", "lib64", "opt", "root", "srv"}
	for _, e := range entries {
		src := "/" + e
		if _, err := os.Stat(src); err == nil {
			tasks = append(tasks, OATask{
				Command:  "oa_bind",
				Src:      src,
				Dst:      filepath.Join(liveroot, e),
				ReadOnly: true,
				Info:     "Bind mount proiettivo: " + e,
			})
		}
	}

	// 4. OVERLAY PER USR E VAR: Le parti che devono essere scrivibili durante il chroot
	for _, ovlDir := range []string{"usr", "var"} {
		lower := filepath.Join(overlay, "lowerdir", ovlDir)
		upper := filepath.Join(overlay, "upperdir", ovlDir)
		work := filepath.Join(overlay, "workdir", ovlDir)
		merged := filepath.Join(liveroot, ovlDir)

		// Creiamo i rami dell'overlay
		tasks = append(tasks, OATask{Command: "oa_mkdir", Path: lower})
		tasks = append(tasks, OATask{Command: "oa_mkdir", Path: upper})
		tasks = append(tasks, OATask{Command: "oa_mkdir", Path: work})

		// Bind della directory originale su lower (ReadOnly)
		tasks = append(tasks, OATask{Command: "oa_bind", Src: "/" + ovlDir, Dst: lower, ReadOnly: true})

		// Mount Overlay finale (unisce lower, upper e work in merged)
		opts := "lowerdir=" + lower + ",upperdir=" + upper + ",workdir=" + work
		tasks = append(tasks, OATask{
			Command: "oa_mount_generic",
			Type:    "overlay",
			Src:     "overlay",
			Dst:     merged,
			Opts:    opts,
			Info:    "Overlay mount per scrivibilità: " + ovlDir,
		})
	}

	// 5. API FILESYSTEMS: Mount dei filesystem virtuali necessari per i processi (chroot)
	tasks = append(tasks, OATask{Command: "oa_mount_generic", Type: "proc", Src: "proc", Dst: filepath.Join(liveroot, "proc")})
	tasks = append(tasks, OATask{Command: "oa_mount_generic", Type: "sysfs", Src: "sys", Dst: filepath.Join(liveroot, "sys")})
	tasks = append(tasks, OATask{Command: "oa_bind", Src: "/dev", Dst: filepath.Join(liveroot, "dev"), Info: "API FS: dev"})
	tasks = append(tasks, OATask{Command: "oa_bind", Src: "/run", Dst: filepath.Join(liveroot, "run"), Info: "API FS: run"})

	return tasks
}
