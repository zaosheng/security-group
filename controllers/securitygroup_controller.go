/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/antihax/optional"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"paas.unicom.cn/dcs-sdk/dcsapi"
	"paas.unicom.cn/dcs-sdk/dcsapi/model/securitygroup"
	"reflect"
	"security-group/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"

	paasv1 "security-group/api/v1"
)

// SecurityGroupReconciler reconciles a SecurityGroup object
type SecurityGroupReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type SecurityGroup struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// +kubebuilder:rbac:groups=paas.unicom.cn,resources=securitygroups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=paas.unicom.cn,resources=securitygroups/status,verbs=get;update;patch

const (
	SecurityGroupFinalizer string = "securitygroup.finalizers.paas.unicom.cn"
)

var config = dcsapi.NewConfigurationWithBasePath("http://172.31.248.3:30086")
var c = dcsapi.NewAPIClient(config)

func (r *SecurityGroupReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("securitygroup", req.NamespacedName)

	sg := &paasv1.SecurityGroup{}
	if err := r.Get(ctx, req.NamespacedName, sg); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			log.Info("没有找到对应的SecurityGroup resource")
			return ctrl.Result{}, nil
		} else {
			log.Error(err, "不是未找到的错误，直接返回错误")
			return ctrl.Result{}, err
		}
	}

	if sg.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("进入 apply SecurityGroup CR 逻辑")
		// 确保 resource 的 finalizers 里有控制器指定的 finalizer
		if !util.ContainsString(sg.ObjectMeta.Finalizers, SecurityGroupFinalizer) {
			log.Info("给 SecurityGroup CR 添加 SecurityGroupFinalizer")
			sg.ObjectMeta.Finalizers = append(sg.ObjectMeta.Finalizers, SecurityGroupFinalizer)
			if err := r.Update(ctx, sg); err != nil {
				return ctrl.Result{}, err
			}
		}
		if _, err := r.applySecurityGroup(ctx, req, sg); err != nil {
			log.Error(err, "apply SecurityGroup CR 失败")
			return ctrl.Result{}, nil
		}
	} else {
		log.Info("进入删除 SecurityGroup CR 的逻辑")
		if util.ContainsString(sg.ObjectMeta.Finalizers, SecurityGroupFinalizer) {
			// 如果 finalizers 被清空，则该 SecurityGroup CR 就已经不存在了，所以必须在次之前删除 SecurityGroup
			log.Info("用sdk删除 SecurityGroup")
			if err := r.cleanSecurityGroup(ctx, req, sg); err != nil {
				log.Error(err, "删除 SecurityGroup CR 失败")
				return ctrl.Result{}, nil
			}
		}
		log.Info("清空 SecurityGroup CR 的 finalizers，SecurityGroup CR 彻底删除")
		sg.ObjectMeta.Finalizers = util.RemoveString(sg.ObjectMeta.Finalizers, SecurityGroupFinalizer)
		if err := r.Update(ctx, sg); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *SecurityGroupReconciler) applySecurityGroup(ctx context.Context, req ctrl.Request, sg *paasv1.SecurityGroup) (*SecurityGroup, error) {
	oldSecurityGroup := &SecurityGroup{}

	// 生成新安全组
	newSecurityGroup := &SecurityGroup{
		Name:        sg.Spec.Name,
		Description: sg.Spec.Description}

	// 安全组存在，更新安全组
	if sg.Status.Id != "" {
		// fmt.Println("------安全组id存在，更新安全组")
		condition_update := paasv1.SpecificationChanging()
		sg.Status.SetConditions(condition_update)
		getSecuritygroupsResponse, _, err := c.SecuritygroupApi.V2SecurityGroupsGet(nil, &dcsapi.SecuritygroupApiV2SecurityGroupsGetOpts{
			XAccountID: optional.NewString(sg.Spec.AccountId),
			XUserID:    optional.NewString(sg.Spec.UserId),
			SearchById: optional.NewString(sg.Status.Id)})
		// 获取安全组失败
		if getSecuritygroupsResponse.Code != 200 {
			// fmt.Println("------安全组id存在,获取安全组失败")
			err := fmt.Errorf("failed to get Securitygroup when updating: %v, %v", getSecuritygroupsResponse.Message, err)
			reconcile_update := paasv1.ReconcileError(err)
			sg.Status.SetConditions(reconcile_update)
			r.Update(ctx, sg)
			return nil, err
		}
		oldSecurityGroup.Name = getSecuritygroupsResponse.Result.List[0].Name
		oldSecurityGroup.Description = getSecuritygroupsResponse.Result.List[0].Description

		// 对比安全组
		if reflect.DeepEqual(oldSecurityGroup, newSecurityGroup) {
			// fmt.Println("------安全组期望状态与实际状态一致，无需更新")
			return oldSecurityGroup, nil
		}

		// 更新安全组
		// 更新状态为修改中
		r.Update(ctx, sg)
		updateSecuritygroupsResponse, _, e := c.SecuritygroupApi.V2SecurityGroupsIdPut(nil, sg.Status.Id, &dcsapi.SecuritygroupApiV2SecurityGroupsIdPutOpts{
			XAccountID: optional.NewString(sg.Spec.AccountId),
			XUserID:    optional.NewString(sg.Spec.UserId),
			Root:       &securitygroup.UpdateSecuritygroupRequest{Name: newSecurityGroup.Name, Description: newSecurityGroup.Description}})
		if updateSecuritygroupsResponse.Code != 200 {
			// 更新安全组失败，更新状态
			// fmt.Printf("------更新安全组失败: %+v\n", updateSecuritygroupsResponse)
			err := fmt.Errorf("failed to update Securitygroup: %+v, %+v", updateSecuritygroupsResponse.Message, e)
			reconcile_update := paasv1.ReconcileError(err)
			sg.Status.SetConditions(reconcile_update)
			r.Update(ctx, sg)
			return oldSecurityGroup, err
		}
		// 更新安全组成功，更新状态为avialable
		// fmt.Printf("------更新安全组成功: %+v\n", updateSecuritygroupsResponse)
		condition_update = paasv1.Available()
		condition_update = condition_update.WithMessage(updateSecuritygroupsResponse.Message)
		reconcile_update := paasv1.ReconcileSuccess()
		sg.Status.SetConditions(condition_update, reconcile_update)
		r.Update(ctx, sg)
		return newSecurityGroup, nil
	}
	// 安全组不存在，创建安全组
	// fmt.Println("------安全组id不存在，创建安全组")
	// 更新状态为创建中
	condition_create := paasv1.Creating()
	sg.Status.SetConditions(condition_create)
	r.Update(ctx, sg)
	createSecuritygroupResponse, _, e := c.SecuritygroupApi.V2SecurityGroupsPost(nil, &dcsapi.SecuritygroupApiV2SecurityGroupsPostOpts{
		XAccountID: optional.NewString(sg.Spec.AccountId),
		XUserID:    optional.NewString(sg.Spec.UserId),
		Root:       &securitygroup.CreateSecuritygroupRequest{Name: newSecurityGroup.Name, Description: newSecurityGroup.Description}})
	if createSecuritygroupResponse.Code != 200 {
		// 创建安全组失败，更新状态
		// fmt.Printf("------创建安全组失败: %+v\n", createSecuritygroupResponse)
		err := fmt.Errorf("failed to create Securitygroup: %+v, %+v", createSecuritygroupResponse.Message, e)
		reconcile_create := paasv1.ReconcileError(err)
		sg.Status.SetConditions(reconcile_create)
		r.Update(ctx, sg)
		return nil, err
	}
	//创建成功，更新状态为avialable
	// fmt.Printf("------创建安全组成功: %+v\n", createSecuritygroupResponse.Result)
	condition_create = paasv1.Available()
	condition_create = condition_create.WithMessage(createSecuritygroupResponse.Message)
	reconcile_create := paasv1.ReconcileSuccess()
	sg.Status.SetConditions(condition_create, reconcile_create)
	r.Update(ctx, sg)
	sg.Status.Id = strconv.FormatInt(createSecuritygroupResponse.Result.Id, 10)
	r.Update(ctx, sg)
	return newSecurityGroup, nil
}

func (r *SecurityGroupReconciler) cleanSecurityGroup(ctx context.Context, req ctrl.Request, sg *paasv1.SecurityGroup) error {
	// 安全组不存在，直接返回
	if sg.Status.Id == "" {
		// fmt.Println("------安全组id不存在，直接返回")
		return nil
	}
	// 更新状态为删除中
	condition_delete := paasv1.Deleting()
	sg.Status.SetConditions(condition_delete)
	r.Update(ctx, sg)
	getSecuritygroupsResponse, _, e := c.SecuritygroupApi.V2SecurityGroupsGet(nil, &dcsapi.SecuritygroupApiV2SecurityGroupsGetOpts{
		XAccountID: optional.NewString(sg.Spec.AccountId),
		XUserID:    optional.NewString(sg.Spec.UserId),
		SearchById: optional.NewString(sg.Status.Id)})
	// 获取安全组失败
	if getSecuritygroupsResponse.Code != 200 {
		// fmt.Println("------删除安全组时获取安全组失败，返回err")
		// 更新删除时状态的message
		err := fmt.Errorf("failed to get Securitygroup when deleting:%+v, %+v", getSecuritygroupsResponse.Message, e)
		reconcile_delete := paasv1.ReconcileError(err)
		sg.Status.SetConditions(reconcile_delete)
		r.Update(ctx, sg)
		return err
	}
	// 安全组不存在，直接返回
	if len(getSecuritygroupsResponse.Result.List) == 0 {
		// fmt.Println("------删除安全组时获取安全组不存在，直接返回")
		return nil
	}
	// 删除安全组
	deleteSecuritygroupResponse, _, e := c.SecuritygroupApi.V2SecurityGroupsIdDelete(nil, sg.Status.Id, &dcsapi.SecuritygroupApiV2SecurityGroupsIdDeleteOpts{
		XAccountID: optional.NewString(sg.Spec.AccountId),
		XUserID:    optional.NewString(sg.Spec.UserId)})
	if deleteSecuritygroupResponse.Code != 200 {
		// fmt.Printf("------删除安全组失败: %+v\n", deleteSecuritygroupResponse)
		err := fmt.Errorf("failed to delete Securitygroup: %+v, %+v", deleteSecuritygroupResponse.Message, e)
		// 更新状态
		reconcile_delete := paasv1.ReconcileError(err)
		sg.Status.SetConditions(reconcile_delete)
		r.Update(ctx, sg)
		return err
	}
	return nil
}

func (r *SecurityGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&paasv1.SecurityGroup{}).
		Complete(r)
}
