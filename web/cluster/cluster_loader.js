export async function loadCluster(indexPath, device, loadWeightsGPU) {
  const index = await fetch(indexPath).then((r) => r.json());
  const promises = index.shards.map((shard) => loadWeightsGPU(device, shard.file));
  return Promise.all(promises);
}
