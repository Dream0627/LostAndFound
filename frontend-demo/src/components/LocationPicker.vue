<!--
  校园地点选择组件（Vue3 <script setup>）。
  能力：
    1) 一键自动定位：浏览器 Geolocation 拿经纬度 → 调后端匹配最近校园预设地点；
    2) 手动兜底：从后端返回的“校区 → 地点”列表中选择（详细到楼/店面）；
    3) 补充说明：选择/定位后可再填一句（如“图书馆东门台阶旁”）。
  对外：emit("confirm", payload)，payload 为后端 /geo/locate 的结果，父组件据此随帖子提交。
-->
<template>
  <div class="location-picker">
    <div class="lp-row">
      <button type="button" class="btn" :disabled="locating" @click="handleGeolocate">
        {{ locating ? "定位中…" : "一键自动定位" }}
      </button>
      <span v-if="matched" class="lp-hint">
        已匹配：{{ matched.location?.name }}
        <template v-if="matched.match_type === 'auto' && matched.distance_meters != null">
          （约 {{ Math.round(matched.distance_meters) }} 米）
        </template>
      </span>
    </div>

    <div v-if="errorMsg" class="alert alert-error mt-8">{{ errorMsg }}</div>

    <div class="field mt-8">
      <label>或手动选择地点</label>
      <select v-model="selectedId" class="select" @change="onManualSelect">
        <option value="">请选择校园地点</option>
        <optgroup v-for="group in groups" :key="group.campus" :label="group.campus">
          <option v-for="loc in group.locations" :key="loc.id" :value="loc.id">
            {{ loc.name }}（{{ loc.category }}）
          </option>
        </optgroup>
      </select>
    </div>

    <div class="field">
      <label>补充说明（可选，≤200 字）</label>
      <input
        v-model="supplement"
        class="input"
        maxlength="200"
        placeholder="如：图书馆东门台阶旁、3 号楼门口快递柜"
      />
    </div>

    <div class="lp-actions">
      <button type="button" class="btn" :disabled="!canConfirm" @click="handleConfirm">
        确认地点
      </button>
      <button v-if="matched || selectedId" type="button" class="btn btn-ghost" @click="handleClear">
        清除
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { fetchCampusLocations, locate } from "@/api/geo";
import { useToastStore } from "@/stores/toast";

const emit = defineEmits(["confirm"]);
const toast = useToastStore();

const groups = ref([]); // 后端返回的“校区 → 地点”分组
const selectedId = ref(""); // 手动选择的地点 ID
const supplement = ref(""); // 补充说明
const matched = ref(null); // 最近一次自动定位/匹配结果
const locating = ref(false); // 是否正在自动定位
const errorMsg = ref(""); // 错误提示

const canConfirm = computed(() => Boolean(matched.value) || Boolean(selectedId.value));

onMounted(async () => {
  try {
    groups.value = await fetchCampusLocations();
  } catch (e) {
    errorMsg.value = "加载预设地点失败，请稍后重试";
  }
});

// 把浏览器 Geolocation 的错误码翻译成可操作的提示，避免“一律显示未授权”误导用户。
// code 约定：1=PERMISSION_DENIED（用户/系统拒绝），2=POSITION_UNAVAILABLE（定位源不可用，如系统定位未开或网络定位不可达），3=TIMEOUT（超时）。
function geoErrorMessage(err) {
  switch (err && err.code) {
    case 1:
      return "定位权限被拒绝：请在浏览器地址栏左侧把本站“位置”权限改为“允许”，并确认系统定位服务已开启";
    case 2:
      return "无法获取位置：请检查系统定位服务是否开启，或当前网络/环境不支持定位（可改用手动选择地点）";
    case 3:
      return "定位超时：请重试，或改用手动选择地点";
    default:
      return "定位失败，请手动选择地点";
  }
}

// 一键自动定位：拿浏览器经纬度交给后端匹配最近地点；失败则引导手动选择。
function handleGeolocate() {
  errorMsg.value = "";
  if (!("geolocation" in navigator)) {
    errorMsg.value = "当前浏览器不支持定位，请手动选择地点";
    return;
  }
  locating.value = true;
  navigator.geolocation.getCurrentPosition(
    async (pos) => {
      try {
        matched.value = await locate({
          latitude: pos.coords.latitude,
          longitude: pos.coords.longitude,
          supplement: supplement.value,
        });
        selectedId.value = "";
      } catch (e) {
        errorMsg.value = e?.msg || "定位匹配失败，请手动选择地点";
      } finally {
        locating.value = false;
      }
    },
    (err) => {
      locating.value = false;
      errorMsg.value = geoErrorMessage(err);
    },
    { enableHighAccuracy: false, timeout: 15000, maximumAge: 60000 }
  );
}

function onManualSelect() {
  // 手动选择后清掉自动定位结果，避免二者混淆。
  if (selectedId.value) matched.value = null;
}

function handleClear() {
  matched.value = null;
  selectedId.value = "";
  errorMsg.value = "";
}

// 确认：把最终地点与补充说明交给父组件。
async function handleConfirm() {
  errorMsg.value = "";
  try {
    const locationId = selectedId.value || matched.value?.location?.id || "";
    const payload = await locate({ locationId, supplement: supplement.value });
    toast.success(`已选择地点：${payload.location?.name || ""}`);
    emit("confirm", payload);
  } catch (e) {
    errorMsg.value = e?.msg || "确认地点失败，请重试";
  }
}
</script>

<style scoped>
.location-picker {
  border: 1px dashed var(--color-border);
  border-radius: var(--radius);
  padding: 16px;
  background: #fbfcfd;
}
.lp-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.lp-hint {
  font-size: 13px;
  color: var(--color-primary-dark);
}
.lp-actions {
  display: flex;
  gap: 10px;
  margin-top: 6px;
}
</style>
