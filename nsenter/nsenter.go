package nsenter

/*
#define _GNU_SOURCE
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>

// 包被引用时，此函数执行
__attribute__((constructor)) void enter_namespace(void) {
	char *mini_docker_pid;
	mini_docker_pid = getenv("mini_docker_pid");
	if (mini_docker_pid) {
		fprintf(stdout, "Got mini_docker_pid=%s\n", mini_docker_pid);
	} else {
		// fprintf(stdout, "Missing mini_docker_pid env skip nsenter");
		return;
	}
	char *mini_docker_cmd;
	mini_docker_cmd = getenv("mini_docker_cmd");
	if (mini_docker_cmd) {
		fprintf(stdout, "got mini_docker_cmd=%s\n", mini_docker_cmd);
	} else {
		// fprintf(stdout, "missing mini_docker_cmd env skip nsenter");
		return;
	}
	int i;
	char nspath[1024];
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };

	for (i=0; i<5; i++) {
		sprintf(nspath, "/proc/%s/ns/%s", mini_docker_pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);

		if (setns(fd, 0) == -1) {
			fprintf(stderr, "Setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
		} else {
			fprintf(stdout, "Setns on %s namespace succeeded\n", namespaces[i]);
		}
		close(fd);
	}
	int res = system(mini_docker_cmd);
	exit(0);
	return;
}
*/
import "C"
