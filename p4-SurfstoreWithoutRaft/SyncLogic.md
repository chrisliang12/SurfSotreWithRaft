# Client Sync Logic
1. Client should first scan the base DIR, and for each file, compute that file's hash list.
2. The client should consult the local index file and compare the results, to see whether
   1. there are now new files in the base dir that aren't in the index file
   2. files that are in the index file, but have changed since the last time the client was executed (hash list is different)
3. Connect to the server and download an updated FileInfoMap (remote Index File)
4. Compare the local index with the remote index:
   1. remote index refers to a file not present in the local index or in the base dir. 
      - should download the blocks associated with that file, reconstitute that file in the base directory, and then add the updated FIleInfo information to the local index
   2. There are new files in the local base dir that are not in the local index or in the remote index.
      - should upload the blocks corresponding to this file to the server, then update the server with the new FileInfo. 
      - if the update is successful, then the client should update its local index.
      - if the update fails, the client should handle this conflict.

# Handling Conflicts
1. if local index and local file are matched but there are new version file on the cloud.
   - download any needed blocks from the server to bring the local file up to the newest version.
   - update the local index, bringing that entry to newest version.
2. when the local index does not matches the local file, and the local index matches the cloud index:
   - update the mapping on the server
   - if success, update the entry in the local index
3. local index does not match the local file, and also does not match remote index:
   - follow the rule that whoever syncs to the cloud first wins.
   - download any required blocks and bring our local version of the file up the date with the cloud version. (throw away the uncommitted local changes)