# Static Provisioning

## Prerequisites

1. Kubernetes 1.13+ (CSI 1.0).
2. The [aws-hyperv-csi-driver](https://github.com/kubernetes-sigs/aws-hyperv-csi-driver) installed.
3. Created an [Amazon EBS volume](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/hyperv-volume-types.html).

## Usage

This example shows you how to create and consume a `PersistentVolume` from an existing EBS volume with static provisioning.

1. Edit the `PersistentVolume` manifest in [pv.yaml](./manifests/pv.yaml) to include your `volumeHandle` EBS volume ID and `nodeSelectorTerms` zone value.
    
    The `StorageClass` on the `PersistentVolumeClaim` and `PersistentVolume` must match. If you have a default storage class, this means you must explicitly set `spec.storageClassName` to `""` in the [PVC manifest](manifests/claim.yaml#L6) if the PV doesn't have a `StorageClass`.
    
    The [`spec.volumeName` field](manifests/claim.yaml#L7) of the PVC must match the [name of the PV](manifests/pv.yaml#L4) for it to be selected.

    ```yaml
    apiVersion: v1
    kind: PersistentVolume
    metadata:
      name: test-pv
    spec:
      accessModes:
      - ReadWriteOnce
      capacity:
        storage: 5Gi
      csi:
        driver: hyperv.csi.k8s.io
        fsType: ext4
        volumeHandle: {EBS volume ID}
      nodeAffinity:
        required:
          nodeSelectorTerms:
            - matchExpressions:
                - key: topology.kubernetes.io/zone
                  operator: In
                  values:
                    - {availability zone}
    ```

2. Deploy the provided pod on your cluster along with the `PersistentVolume` and `PersistentVolumeClaim`:
    ```sh
    $ kubectl apply -f manifests

    persistentvolumeclaim/hyperv-claim created
    pod/app created
    persistentvolume/test-pv created
    ```

3. Validate the `PersistentVolumeClaim` is bound to your `PersistentVolume`.
    ```sh
    $ kubectl get pvc hyperv-claim

    NAME        STATUS   VOLUME    CAPACITY   ACCESS MODES   STORAGECLASS   AGE
    hyperv-claim   Bound    test-pv   5Gi        RWO                           53s
    ```

4. Validate the pod successfully wrote data to the statically provisioned volume:
    ```sh
    $ kubectl exec app -- cat /data/out.txt

    Tue Feb 22 20:51:37 UTC 2022
    ...
    ```

5. Cleanup resources:
    ```sh
    $ kubectl delete -f manifests

    persistentvolumeclaim "hyperv-claim" deleted
    pod "app" deleted
    persistentvolume "test-pv" deleted
    ```
