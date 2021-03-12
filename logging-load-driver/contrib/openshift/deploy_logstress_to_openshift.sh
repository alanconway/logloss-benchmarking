#!/bin/bash

source ./deploy_functions.sh

main() {
  select_node_to_use
  configure_workers_log_rotation
  return_node_to_normal
  delete_logstress_project_if_exists
  create_logstress_project
  set_credentials
  deploy_logstress
  deploy_log_collector
  deploy_capture_statistics
  evacuate_node_for_performance_tests
  print_pods_status
  print_usage_instructions
}

main

